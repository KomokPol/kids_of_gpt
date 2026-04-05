from __future__ import annotations

import os
from contextlib import asynccontextmanager

import redis.asyncio as aioredis
import structlog
from fastapi import FastAPI

structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ],
)
log = structlog.get_logger()

app_state: dict = {}


@asynccontextmanager
async def lifespan(app_inst: FastAPI):
    redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
    log.info("notification.startup", redis_url=redis_url)
    try:
        r = aioredis.from_url(redis_url, decode_responses=False)
        await r.ping()
        app_state["redis"] = r
        from .storage import NotificationStorage
        app_state["storage"] = NotificationStorage(r)
        log.info("notification.redis_connected")
    except Exception:
        log.error("notification.redis_connect_failed", redis_url=redis_url)
        raise
    yield
    r = app_state.get("redis")
    if r:
        await r.aclose()
    log.info("notification.shutdown")


app = FastAPI(title="ZONDEX Notification Service", version="1.0.0", lifespan=lifespan)

from .routes import router  # noqa: E402

app.include_router(router)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "notification"}
