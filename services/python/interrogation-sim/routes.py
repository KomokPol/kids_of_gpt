from __future__ import annotations

from fastapi import APIRouter, HTTPException

from models import AnswerRequest, AnswerResult, ScenarioListItem, SessionOut, StartSessionRequest
from scenarios import SCENARIOS

router = APIRouter()


def get_engine():
    from main import app_state

    return app_state["engine"]


@router.get("/scenarios", response_model=list[ScenarioListItem])
async def list_scenarios() -> list[ScenarioListItem]:
    return [
        ScenarioListItem(
            id=s.id,
            title=s.title,
            description=s.description,
            question_count=len(s.questions),
        )
        for s in SCENARIOS
    ]


@router.post("/sessions", response_model=SessionOut)
async def create_session(body: StartSessionRequest) -> SessionOut:
    eng = get_engine()
    try:
        return eng.start_session(body.user_id, body.scenario_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e)) from e


@router.post("/sessions/{session_id}/answer", response_model=AnswerResult)
async def submit_answer(session_id: str, body: AnswerRequest) -> AnswerResult:
    eng = get_engine()
    try:
        correct, explanation, score_delta, session = eng.answer(session_id, body.answer)
    except KeyError as e:
        raise HTTPException(status_code=404, detail="Сессия не найдена") from e
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e)) from e

    return AnswerResult(
        correct=correct,
        explanation=explanation,
        score_delta=score_delta,
        session=session,
    )


@router.get("/sessions/{session_id}", response_model=SessionOut)
async def get_session(session_id: str) -> SessionOut:
    eng = get_engine()
    try:
        return eng.get_session(session_id)
    except KeyError as e:
        raise HTTPException(status_code=404, detail="Сессия не найдена") from e
