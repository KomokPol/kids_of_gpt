from __future__ import annotations

import json
from contextlib import asynccontextmanager
from pathlib import Path

import structlog
from fastapi import FastAPI

from .engine import RecommendationEngine
from .routes import router

structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ],
)
log = structlog.get_logger()

app_state: dict = {}


@asynccontextmanager
async def lifespan(app: FastAPI):
    data_path = Path(__file__).resolve().parent / "data" / "items.json"
    log.info("recommendation.startup", data_path=str(data_path))
    raw = json.loads(data_path.read_text(encoding="utf-8"))
    if not isinstance(raw, list):
        raise ValueError("items.json must contain a JSON array")
    engine = RecommendationEngine(raw)
    app_state["engine"] = engine
    yield
    app_state.clear()


app = FastAPI(
    title="ZONDEX Recommendation Service",
    version="1.0.0",
    lifespan=lifespan,
)
app.include_router(router)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "recommendation"}
