from __future__ import annotations

import os
import sys

import pytest
from fastapi.testclient import TestClient

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if ROOT not in sys.path:
    sys.path.insert(0, ROOT)

import main as main_module
from engine import InterrogationEngine


@pytest.fixture
def client() -> TestClient:
    with TestClient(main_module.app) as c:
        main_module.app_state["engine"] = InterrogationEngine()
        yield c


def test_health(client: TestClient) -> None:
    r = client.get("/health")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"


def test_list_scenarios(client: TestClient) -> None:
    r = client.get("/scenarios")
    assert r.status_code == 200
    data = r.json()
    assert len(data) == 4
    ids = {s["id"] for s in data}
    assert ids == {"transfer_scam", "fake_lawyer", "phone_scam", "authority_pressure"}
    for s in data:
        assert "question_count" in s
        assert s["question_count"] == 4


def test_start_session(client: TestClient) -> None:
    r = client.post("/sessions", json={"user_id": "u1", "scenario_id": "transfer_scam"})
    assert r.status_code == 200
    body = r.json()
    assert body["user_id"] == "u1"
    assert body["scenario_id"] == "transfer_scam"
    assert body["status"] == "active"
    assert body["questions_total"] == 4
    assert body["questions_answered"] == 0
    assert body["current_question"] is not None
    assert body["score"] == 0
    assert body["max_score"] > 0


def test_start_session_unknown_scenario(client: TestClient) -> None:
    r = client.post("/sessions", json={"user_id": "u1", "scenario_id": "nope"})
    assert r.status_code == 400


def test_answer_correct(client: TestClient) -> None:
    s = client.post("/sessions", json={"user_id": "u1", "scenario_id": "transfer_scam"}).json()
    sid = s["session_id"]
    r = client.post(
        f"/sessions/{sid}/answer",
        json={"answer": "Нужно проверить лично и через официальные каналы, не переводить срочно"},
    )
    assert r.status_code == 200
    out = r.json()
    assert out["correct"] is True
    assert out["score_delta"] > 0
    assert out["session"]["questions_answered"] == 1
    assert out["session"]["score"] == out["score_delta"]


def test_answer_incorrect(client: TestClient) -> None:
    s = client.post("/sessions", json={"user_id": "u1", "scenario_id": "transfer_scam"}).json()
    sid = s["session_id"]
    r = client.post(
        f"/sessions/{sid}/answer",
        json={"answer": "Сразу переведу всё, что попросят, без вопросов"},
    )
    assert r.status_code == 200
    out = r.json()
    assert out["correct"] is False
    assert out["score_delta"] == 0
    assert out["session"]["questions_answered"] == 1
    assert out["session"]["score"] == 0


def test_complete_full_scenario(client: TestClient) -> None:
    s = client.post("/sessions", json={"user_id": "u2", "scenario_id": "phone_scam"}).json()
    sid = s["session_id"]
    answers = [
        "Не называю код из смс, сам перезвоню в официальный банк",
        "Сначала проверю через родственников и другой номер, не перевожу",
        "Нет, удалённый доступ опасно, не устанавливаю",
        "Проверить правила через администрацию официально, это типичная схема",
    ]
    total_score = 0
    for a in answers:
        r = client.post(f"/sessions/{sid}/answer", json={"answer": a})
        assert r.status_code == 200
        body = r.json()
        assert body["correct"] is True
        total_score += body["score_delta"]

    final = r.json()["session"]
    assert final["status"] == "completed"
    assert final["current_question"] is None
    assert final["questions_answered"] == 4
    assert final["score"] == total_score
    assert final["xp_reward"] == final["score"] * 2
    assert final["money_reward"] == final["score"] * 5


def test_get_session_status(client: TestClient) -> None:
    s = client.post("/sessions", json={"user_id": "u3", "scenario_id": "fake_lawyer"}).json()
    sid = s["session_id"]
    r = client.get(f"/sessions/{sid}")
    assert r.status_code == 200
    assert r.json()["session_id"] == sid
    assert r.json()["scenario_id"] == "fake_lawyer"


def test_get_session_not_found(client: TestClient) -> None:
    r = client.get("/sessions/00000000-0000-0000-0000-000000000000")
    assert r.status_code == 404


def test_answer_after_completed_returns_400(client: TestClient) -> None:
    s = client.post("/sessions", json={"user_id": "u4", "scenario_id": "authority_pressure"}).json()
    sid = s["session_id"]
    good = [
        "Это манипуляция именем начальника, нужно проверить",
        "Отказ, не плачу за вымогательство, только легально",
        "Это публичный стыд и давление, спокойно не поддаюсь",
        "Заявление в администрацию о шантаже, не соглашаться на угрозы, нужна помощь",
    ]
    for a in good:
        r = client.post(f"/sessions/{sid}/answer", json={"answer": a})
        assert r.status_code == 200

    r2 = client.post(f"/sessions/{sid}/answer", json={"answer": "ещё ответ"})
    assert r2.status_code == 400


def test_evaluate_answer_scoring() -> None:
    from scoring import evaluate_answer

    ok, pts = evaluate_answer("Свяжусь с канцелярия суда для проверки", ["канцелярия"], 10)
    assert ok is True
    assert pts == 10

    ok2, pts2 = evaluate_answer("переведу всё сразу", ["канцелярия"], 10)
    assert ok2 is False
    assert pts2 == 0

    ok3, pts3 = evaluate_answer("КАНЦЕЛЯРИЯ", ["канцелярия"], 7)
    assert ok3 is True
    assert pts3 == 7
