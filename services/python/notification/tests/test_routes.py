from __future__ import annotations

import sys
from pathlib import Path

_PY_ROOT = Path(__file__).resolve().parents[2]
if str(_PY_ROOT) not in sys.path:
    sys.path.insert(0, str(_PY_ROOT))

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient

from notification.main import app, app_state
from notification.storage import NotificationStorage


def _key(name: str | bytes) -> str:
    return name if isinstance(name, str) else name.decode("utf-8")


class FakeRedis:
    """In-memory Redis list emulation for async storage tests."""

    def __init__(self) -> None:
        self._lists: dict[str, list[str]] = {}

    async def lpush(self, name: str | bytes, *values: str | bytes) -> int:
        key = _key(name)
        lst = self._lists.setdefault(key, [])
        for v in values:
            s = v.decode("utf-8") if isinstance(v, bytes) else v
            lst.insert(0, s)
        return len(lst)

    async def expire(self, name: str | bytes, time: int) -> bool:
        return True

    async def lrange(
        self, name: str | bytes, start: int, end: int
    ) -> list[bytes]:
        key = _key(name)
        lst = self._lists.get(key, [])
        if end == -1:
            sl = lst[start:]
        else:
            sl = lst[start : end + 1]
        return [x.encode("utf-8") for x in sl]

    async def llen(self, name: str | bytes) -> int:
        key = _key(name)
        return len(self._lists.get(key, []))

    async def lindex(self, name: str | bytes, index: int) -> bytes | None:
        key = _key(name)
        lst = self._lists.get(key, [])
        if index < 0:
            idx = len(lst) + index
        else:
            idx = index
        if 0 <= idx < len(lst):
            return lst[idx].encode("utf-8")
        return None

    async def lset(
        self, name: str | bytes, index: int, value: str | bytes
    ) -> bool:
        key = _key(name)
        lst = self._lists[key]
        s = value.decode("utf-8") if isinstance(value, bytes) else value
        lst[index] = s
        return True

    async def aclose(self) -> None:
        pass


@pytest_asyncio.fixture
async def client() -> AsyncClient:
    app_state.clear()
    fake = FakeRedis()
    app_state["redis"] = fake
    app_state["storage"] = NotificationStorage(fake)
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac
    app_state.clear()


@pytest.mark.asyncio
async def test_send_list_mark_read_unread_count_health(client: AsyncClient) -> None:
    uid = "user-1"
    create = await client.post(
        "/notifications",
        json={"user_id": uid, "title": "Hello", "body": "World"},
    )
    assert create.status_code == 201
    data = create.json()
    nid = data["notification_id"]
    assert data["user_id"] == uid
    assert data["title"] == "Hello"
    assert data["body"] == "World"
    assert data["read"] is False

    listed = await client.get(f"/notifications/{uid}")
    assert listed.status_code == 200
    body = listed.json()
    assert body["user_id"] == uid
    assert body["total"] == 1
    assert len(body["notifications"]) == 1
    assert body["notifications"][0]["notification_id"] == nid

    unread = await client.get(f"/notifications/{uid}/unread-count")
    assert unread.status_code == 200
    assert unread.json() == {"user_id": uid, "unread_count": 1}

    patch = await client.patch(f"/notifications/{uid}/{nid}/read")
    assert patch.status_code == 200
    assert patch.json() == {"ok": True}

    unread2 = await client.get(f"/notifications/{uid}/unread-count")
    assert unread2.json() == {"user_id": uid, "unread_count": 0}

    listed2 = await client.get(f"/notifications/{uid}")
    assert listed2.json()["notifications"][0]["read"] is True

    health = await client.get("/health")
    assert health.status_code == 200
    assert health.json() == {"status": "ok", "service": "notification"}


@pytest.mark.asyncio
async def test_list_limit_query(client: AsyncClient) -> None:
    uid = "user-2"
    for i in range(3):
        r = await client.post(
            "/notifications",
            json={"user_id": uid, "title": f"t{i}", "body": "b"},
        )
        assert r.status_code == 201
    r = await client.get(f"/notifications/{uid}", params={"limit": 2})
    assert r.status_code == 200
    assert r.json()["total"] == 3
    assert len(r.json()["notifications"]) == 2
