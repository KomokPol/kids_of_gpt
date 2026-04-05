from __future__ import annotations

import os

import redis.asyncio as aioredis
import structlog
from fastapi import FastAPI

from .aggregator import Aggregator
from .routes import router

structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ],
)
log = structlog.get_logger()

app = FastAPI(title="ZONDEX Analytics Service", version="1.0.0")
app.include_router(router)

app_state: dict = {}


@app.on_event("startup")
async def startup() -> None:
    redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
    log.info("analytics.startup", redis_url=redis_url)
    r = aioredis.from_url(redis_url, decode_responses=False)
    app_state["redis"] = r
    app_state["aggregator"] = Aggregator(r)


@app.on_event("shutdown")
async def shutdown() -> None:
    r = app_state.get("redis")
    if r:
        await r.aclose()
    log.info("analytics.shutdown")


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "analytics"}
