from __future__ import annotations

import random
import uuid
from dataclasses import dataclass

from models import QuestionOut, SessionOut, SessionStatus
from scenarios import SCENARIOS, Scenario, ScenarioQuestion
from scoring import evaluate_answer


def _scenario_max_score(scenario: Scenario) -> int:
    return sum(q.points for q in scenario.questions)


def _question_out(q: ScenarioQuestion) -> QuestionOut:
    return QuestionOut(
        question_id=q.id,
        text=q.text,
        hints=list(q.hints),
        time_limit_seconds=60,
    )


@dataclass
class _SessionState:
    session_id: str
    user_id: str
    scenario: Scenario
    current_index: int = 0
    score: int = 0
    questions_answered: int = 0
    status: SessionStatus = SessionStatus.active


class InterrogationEngine:
    def __init__(self) -> None:
        self._sessions: dict[str, _SessionState] = {}
        self._by_scenario_id: dict[str, Scenario] = {s.id: s for s in SCENARIOS}

    def start_session(self, user_id: str, scenario_id: str | None = None) -> SessionOut:
        if scenario_id is None:
            scenario = random.choice(SCENARIOS)
        else:
            scenario = self._by_scenario_id.get(scenario_id)
            if scenario is None:
                raise ValueError(f"Неизвестный сценарий: {scenario_id}")

        sid = str(uuid.uuid4())
        self._sessions[sid] = _SessionState(session_id=sid, user_id=user_id, scenario=scenario)
        return self._to_session_out(self._sessions[sid])

    def answer(self, session_id: str, answer_text: str) -> tuple[bool, str, int, SessionOut]:
        state = self._sessions.get(session_id)
        if state is None:
            raise KeyError(session_id)
        if state.status != SessionStatus.active:
            raise ValueError("Сессия уже завершена")

        questions = state.scenario.questions
        if state.current_index >= len(questions):
            raise ValueError("Нет текущего вопроса")

        q = questions[state.current_index]
        correct, delta = evaluate_answer(answer_text, q.correct_keywords, q.points)
        explanation = q.explanation

        state.score += delta
        state.questions_answered += 1
        state.current_index += 1

        if state.current_index >= len(questions):
            state.status = SessionStatus.completed

        session_out = self._to_session_out(state)
        return correct, explanation, delta, session_out

    def get_session(self, session_id: str) -> SessionOut:
        state = self._sessions.get(session_id)
        if state is None:
            raise KeyError(session_id)
        return self._to_session_out(state)

    def _to_session_out(self, state: _SessionState) -> SessionOut:
        scenario = state.scenario
        total = len(scenario.questions)
        max_score = _scenario_max_score(scenario)
        current_q: ScenarioQuestion | None = None
        if state.status == SessionStatus.active and state.current_index < total:
            current_q = scenario.questions[state.current_index]

        current_out = _question_out(current_q) if current_q is not None else None

        xp_reward = 0
        money_reward = 0
        if state.status == SessionStatus.completed:
            xp_reward = state.score * 2
            money_reward = state.score * 5

        return SessionOut(
            session_id=state.session_id,
            user_id=state.user_id,
            scenario_id=scenario.id,
            scenario_title=scenario.title,
            status=state.status,
            current_question=current_out,
            questions_total=total,
            questions_answered=state.questions_answered,
            score=state.score,
            max_score=max_score,
            xp_reward=xp_reward,
            money_reward=money_reward,
        )
