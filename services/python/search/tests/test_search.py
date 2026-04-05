from __future__ import annotations

import sys
from pathlib import Path

import pytest
from fastapi.testclient import TestClient

# Запуск тестов: из каталога services/python с PYTHONPATH=.
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from search.main import app  # noqa: E402


@pytest.fixture
def client() -> TestClient:
    with TestClient(app) as c:
        yield c


def test_search_finds_items(client: TestClient) -> None:
    r = client.get("/search", params={"q": "сигареты"})
    assert r.status_code == 200
    data = r.json()
    assert data["total"] >= 1
    assert any("сигарет" in it["title"].lower() for it in data["items"])


def test_category_filter(client: TestClient) -> None:
    r = client.get("/search", params={"q": "кино", "category": "kinoshka"})
    assert r.status_code == 200
    data = r.json()
    for it in data["items"]:
        assert it["category"] == "kinoshka"


def test_search_enabled_false_returns_403(client: TestClient) -> None:
    r = client.get("/search", params={"q": "чай", "search_enabled": "false"})
    assert r.status_code == 403
    assert r.json()["detail"] == "Поиск заблокирован. Достигните уровня 3 для разблокировки"


def test_empty_query(client: TestClient) -> None:
    r = client.get("/search", params={"q": ""})
    assert r.status_code == 200
    data = r.json()
    assert data["items"] == []
    assert data["total"] == 0


def test_catalog_endpoint(client: TestClient) -> None:
    r = client.get("/catalog")
    assert r.status_code == 200
    items = r.json()
    assert len(items) == 30
