from __future__ import annotations

import json
from typing import TYPE_CHECKING

from .models import NotificationIn, NotificationStored

if TYPE_CHECKING:
    from redis.asyncio import Redis

KEY_PREFIX = "notifications:"
TTL_SECONDS = 86400


class NotificationStorage:
    def __init__(self, redis: Redis) -> None:
        self._redis = redis

    def _key(self, user_id: str) -> str:
        return f"{KEY_PREFIX}{user_id}"

    async def save(self, notification: NotificationIn) -> NotificationStored:
        stored = NotificationStored(**notification.model_dump())
        key = self._key(stored.user_id)
        payload = json.dumps(stored.model_dump(), ensure_ascii=False)
        await self._redis.lpush(key, payload)
        await self._redis.expire(key, TTL_SECONDS)
        return stored

    async def list(
        self, user_id: str, limit: int = 50
    ) -> list[NotificationStored]:
        if limit <= 0:
            return []
        key = self._key(user_id)
        raw_items = await self._redis.lrange(key, 0, limit - 1)
        out: list[NotificationStored] = []
        for raw in raw_items:
            text = raw.decode("utf-8") if isinstance(raw, bytes) else raw
            data = json.loads(text)
            out.append(NotificationStored(**data))
        return out

    async def mark_read(self, user_id: str, notification_id: str) -> bool:
        key = self._key(user_id)
        n = await self._redis.llen(key)
        for i in range(n):
            raw = await self._redis.lindex(key, i)
            if raw is None:
                continue
            text = raw.decode("utf-8") if isinstance(raw, bytes) else raw
            data = json.loads(text)
            if data.get("notification_id") == notification_id:
                data["read"] = True
                new_payload = json.dumps(data, ensure_ascii=False)
                await self._redis.lset(key, i, new_payload)
                await self._redis.expire(key, TTL_SECONDS)
                return True
        return False

    async def count_unread(self, user_id: str) -> int:
        key = self._key(user_id)
        n = await self._redis.llen(key)
        count = 0
        for i in range(n):
            raw = await self._redis.lindex(key, i)
            if raw is None:
                continue
            text = raw.decode("utf-8") if isinstance(raw, bytes) else raw
            data = json.loads(text)
            if not data.get("read", False):
                count += 1
        return count

    async def count_total(self, user_id: str) -> int:
        return await self._redis.llen(self._key(user_id))
