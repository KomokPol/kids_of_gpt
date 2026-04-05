from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from recommendation.main import app


@pytest.fixture
def client() -> TestClient:
    with TestClient(app) as c:
        yield c


def test_recommendations_level_0_excludes_high_min_level(client: TestClient) -> None:
    resp = client.get("/recommendations/user-l0", params={"level": 0})
    assert resp.status_code == 200
    data = resp.json()
    assert data["user_id"] == "user-l0"
    assert data["level"] == 0
    ids = {x["item_id"] for x in data["recommendations"]}
    assert "kinoshka-igla" not in ids
    assert "barygi-ruchka-snaiperskaya" not in ids
    assert all(x["min_level"] <= 0 for x in data["recommendations"])


def test_recommendations_level_5_includes_all(client: TestClient) -> None:
    catalog = client.get("/items")
    assert catalog.status_code == 200
    all_ids = {x["item_id"] for x in catalog.json()}

    resp = client.get("/recommendations/user-l5", params={"level": 5, "limit": 100})
    assert resp.status_code == 200
    data = resp.json()
    rec_ids = {x["item_id"] for x in data["recommendations"]}
    assert rec_ids == all_ids


def test_recommendations_filter_by_category(client: TestClient) -> None:
    resp = client.get(
        "/recommendations/user-cat",
        params={"level": 5, "category": "balanda", "limit": 50},
    )
    assert resp.status_code == 200
    data = resp.json()
    assert data["total"] >= 1
    assert all(x["category"] == "balanda" for x in data["recommendations"])


def test_recommendations_unknown_user_ok(client: TestClient) -> None:
    resp = client.get("/recommendations/unknown-user-xyz", params={"level": 1})
    assert resp.status_code == 200
    data = resp.json()
    assert data["user_id"] == "unknown-user-xyz"
    assert isinstance(data["recommendations"], list)
    assert data["total"] == len(data["recommendations"])


def test_items_and_item_detail(client: TestClient) -> None:
    lst = client.get("/items")
    assert lst.status_code == 200
    first = lst.json()[0]
    rid = first["item_id"]
    one = client.get(f"/items/{rid}")
    assert one.status_code == 200
    assert one.json()["item_id"] == rid

    missing = client.get("/items/does-not-exist")
    assert missing.status_code == 404
