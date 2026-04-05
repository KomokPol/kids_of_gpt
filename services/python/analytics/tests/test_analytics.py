import pytest
from httpx import ASGITransport, AsyncClient

from analytics.main import app, app_state
from analytics.aggregator import Aggregator


class FakeRedis:
    """Minimal in-memory Redis mock for testing."""

    def __init__(self):
        self._data: dict = {}
        self._lists: dict[str, list] = {}
        self._hashes: dict[str, dict] = {}
        self._sorted_sets: dict[str, dict] = {}
        self._sets: dict[str, set] = {}

    def pipeline(self, transaction=False):
        return FakePipeline(self)

    async def get(self, key):
        return self._data.get(key)

    async def hgetall(self, key):
        h = self._hashes.get(key, {})
        return {k.encode(): v.encode() for k, v in h.items()}

    async def zrevrange(self, key, start, end, withscores=False):
        ss = self._sorted_sets.get(key, {})
        items = sorted(ss.items(), key=lambda x: x[1], reverse=True)
        sliced = items[start : end + 1]
        if withscores:
            return [(k.encode(), v) for k, v in sliced]
        return [k.encode() for k, _ in sliced]

    async def zcard(self, key):
        return len(self._sorted_sets.get(key, {}))

    async def lrange(self, key, start, end):
        lst = self._lists.get(key, [])
        return [v.encode() if isinstance(v, str) else v for v in lst[start : end + 1]]

    async def smembers(self, key):
        s = self._sets.get(key, set())
        return {v.encode() for v in s}

    async def aclose(self):
        pass


class FakePipeline:
    def __init__(self, r: FakeRedis):
        self._r = r
        self._ops: list = []

    def incr(self, key):
        self._ops.append(("incr", key))
        return self

    def hincrby(self, key, field, amount):
        self._ops.append(("hincrby", key, field, amount))
        return self

    def hsetnx(self, key, field, value):
        self._ops.append(("hsetnx", key, field, value))
        return self

    def hset(self, key, field, value):
        self._ops.append(("hset", key, field, value))
        return self

    def zincrby(self, key, amount, member):
        self._ops.append(("zincrby", key, amount, member))
        return self

    def lpush(self, key, value):
        self._ops.append(("lpush", key, value))
        return self

    def ltrim(self, key, start, end):
        self._ops.append(("ltrim", key, start, end))
        return self

    def sadd(self, key, *values):
        self._ops.append(("sadd", key, *values))
        return self

    async def execute(self):
        r = self._r
        for op in self._ops:
            cmd = op[0]
            if cmd == "incr":
                val = r._data.get(op[1], b"0")
                if isinstance(val, bytes):
                    val = val.decode()
                r._data[op[1]] = str(int(val) + 1).encode()
            elif cmd == "hincrby":
                h = r._hashes.setdefault(op[1], {})
                h[op[2]] = str(int(h.get(op[2], "0")) + op[3])
            elif cmd == "hsetnx":
                h = r._hashes.setdefault(op[1], {})
                if op[2] not in h:
                    h[op[2]] = op[3]
            elif cmd == "hset":
                h = r._hashes.setdefault(op[1], {})
                h[op[2]] = op[3]
            elif cmd == "zincrby":
                ss = r._sorted_sets.setdefault(op[1], {})
                ss[op[3]] = ss.get(op[3], 0) + op[2]
            elif cmd == "lpush":
                lst = r._lists.setdefault(op[1], [])
                lst.insert(0, op[2])
            elif cmd == "ltrim":
                lst = r._lists.get(op[1], [])
                r._lists[op[1]] = lst[op[2] : op[3] + 1]
            elif cmd == "sadd":
                s = r._sets.setdefault(op[1], set())
                for v in op[2:]:
                    s.add(v)


@pytest.fixture
def fake_redis():
    return FakeRedis()


@pytest.fixture
def setup_app(fake_redis):
    app_state["redis"] = fake_redis
    app_state["aggregator"] = Aggregator(fake_redis)
    return app


@pytest.mark.asyncio
async def test_ingest_event(setup_app):
    transport = ASGITransport(app=setup_app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.post("/events", json={
            "event_type": "burmalda.spin_completed",
            "aggregate_type": "spin",
            "aggregate_id": "spin-1",
            "user_id": "user-1",
        })
        assert resp.status_code == 201
        data = resp.json()
        assert data["event_type"] == "burmalda.spin_completed"
        assert "event_id" in data
        assert "occurred_at" in data


@pytest.mark.asyncio
async def test_user_stats_empty(setup_app):
    transport = ASGITransport(app=setup_app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.get("/stats/unknown-user")
        assert resp.status_code == 200
        data = resp.json()
        assert data["total_events"] == 0
        assert data["streak_days"] == 0


@pytest.mark.asyncio
async def test_user_stats_after_events(setup_app):
    transport = ASGITransport(app=setup_app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        for i in range(3):
            await client.post("/events", json={
                "event_type": "xp.granted",
                "aggregate_type": "progression",
                "aggregate_id": f"xp-{i}",
                "user_id": "user-42",
            })

        resp = await client.get("/stats/user-42")
        assert resp.status_code == 200
        data = resp.json()
        assert data["total_events"] == 3
        assert data["event_counts"]["xp.granted"] == 3


@pytest.mark.asyncio
async def test_dashboard(setup_app):
    transport = ASGITransport(app=setup_app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        await client.post("/events", json={
            "event_type": "wallet.credited",
            "aggregate_type": "wallet",
            "aggregate_id": "w-1",
            "user_id": "user-1",
        })
        resp = await client.get("/dashboard")
        assert resp.status_code == 200
        data = resp.json()
        assert data["total_events"] >= 1
        assert data["total_users"] >= 1


@pytest.mark.asyncio
async def test_health(setup_app):
    transport = ASGITransport(app=setup_app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.get("/health")
        assert resp.status_code == 200
        assert resp.json()["status"] == "ok"
