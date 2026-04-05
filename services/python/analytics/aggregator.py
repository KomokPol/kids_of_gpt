from __future__ import annotations

import json
from datetime import datetime, timezone

import redis.asyncio as aioredis

EVENTS_STREAM = "analytics:events"
USER_EVENTS_PREFIX = "analytics:user:"
EVENT_TYPE_COUNTS = "analytics:event_types"
GLOBAL_COUNTER = "analytics:total"
TOP_USERS = "analytics:top_users"
USER_DAYS_PREFIX = "analytics:days:"

MAX_RECENT = 50


class Aggregator:
    def __init__(self, redis_client: aioredis.Redis) -> None:
        self._r = redis_client

    async def ingest(self, event: dict) -> None:
        user_id = event["user_id"]
        event_type = event["event_type"]
        ts = event.get("occurred_at", datetime.now(timezone.utc).isoformat())
        compact = json.dumps(event, ensure_ascii=False)

        pipe = self._r.pipeline(transaction=False)
        pipe.incr(GLOBAL_COUNTER)
        pipe.hincrby(EVENT_TYPE_COUNTS, event_type, 1)
        pipe.hincrby(f"{USER_EVENTS_PREFIX}{user_id}", event_type, 1)
        pipe.hincrby(f"{USER_EVENTS_PREFIX}{user_id}", "__total__", 1)
        pipe.hsetnx(f"{USER_EVENTS_PREFIX}{user_id}", "__first_seen__", ts)
        pipe.hset(f"{USER_EVENTS_PREFIX}{user_id}", "__last_seen__", ts)
        pipe.zincrby(TOP_USERS, 1, user_id)
        pipe.lpush(EVENTS_STREAM, compact)
        pipe.ltrim(EVENTS_STREAM, 0, MAX_RECENT - 1)

        today = datetime.now(timezone.utc).strftime("%Y-%m-%d")
        pipe.sadd(f"{USER_DAYS_PREFIX}{user_id}", today)

        await pipe.execute()

    async def user_stats(self, user_id: str) -> dict:
        data = await self._r.hgetall(f"{USER_EVENTS_PREFIX}{user_id}")
        if not data:
            return {
                "user_id": user_id,
                "total_events": 0,
                "event_counts": {},
                "first_seen": None,
                "last_seen": None,
                "streak_days": 0,
            }

        decoded = {k.decode(): v.decode() for k, v in data.items()}
        total = int(decoded.pop("__total__", "0"))
        first_seen = decoded.pop("__first_seen__", None)
        last_seen = decoded.pop("__last_seen__", None)
        event_counts = {k: int(v) for k, v in decoded.items()}

        streak = await self._calculate_streak(user_id)

        return {
            "user_id": user_id,
            "total_events": total,
            "event_counts": event_counts,
            "first_seen": first_seen,
            "last_seen": last_seen,
            "streak_days": streak,
        }

    async def dashboard(self) -> dict:
        total_str = await self._r.get(GLOBAL_COUNTER)
        total = int(total_str) if total_str else 0

        raw_types = await self._r.hgetall(EVENT_TYPE_COUNTS)
        event_type_counts = {
            k.decode(): int(v) for k, v in raw_types.items()
        } if raw_types else {}

        raw_top = await self._r.zrevrange(TOP_USERS, 0, 9, withscores=True)
        top_users = [
            {"user_id": uid.decode(), "events": int(score)}
            for uid, score in raw_top
        ]

        total_users = await self._r.zcard(TOP_USERS)

        raw_recent = await self._r.lrange(EVENTS_STREAM, 0, 9)
        recent = [json.loads(e) for e in raw_recent]

        return {
            "total_events": total,
            "total_users": total_users,
            "event_type_counts": event_type_counts,
            "top_users": top_users,
            "recent_events": recent,
        }

    async def _calculate_streak(self, user_id: str) -> int:
        raw_days = await self._r.smembers(f"{USER_DAYS_PREFIX}{user_id}")
        if not raw_days:
            return 0

        days = sorted(
            (datetime.strptime(d.decode(), "%Y-%m-%d").date() for d in raw_days),
            reverse=True,
        )

        streak = 1
        for i in range(1, len(days)):
            if (days[i - 1] - days[i]).days == 1:
                streak += 1
            else:
                break
        return streak
