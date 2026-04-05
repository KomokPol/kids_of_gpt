from __future__ import annotations

from enum import Enum

from pydantic import BaseModel, Field


class SessionStatus(str, Enum):
    active = "active"
    completed = "completed"
    failed = "failed"


class StartSessionRequest(BaseModel):
    user_id: str
    scenario_id: str | None = None  # None = random


class AnswerRequest(BaseModel):
    answer: str


class QuestionOut(BaseModel):
    question_id: str
    text: str
    hints: list[str] = Field(default_factory=list)
    time_limit_seconds: int = 60


class SessionOut(BaseModel):
    session_id: str
    user_id: str
    scenario_id: str
    scenario_title: str
    status: SessionStatus
    current_question: QuestionOut | None = None
    questions_total: int = 0
    questions_answered: int = 0
    score: int = 0
    max_score: int = 0
    xp_reward: int = 0
    money_reward: int = 0


class AnswerResult(BaseModel):
    correct: bool
    explanation: str
    score_delta: int
    session: SessionOut


class ScenarioListItem(BaseModel):
    id: str
    title: str
    description: str
    question_count: int
