from __future__ import annotations

import os

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

app = FastAPI(title="ZONDEX Notification Service", version="1.0.0")

from .routes import router  # noqa: E402

app.include_router(router)


@app.on_event("startup")
async def startup() -> None:
    if app_state.get("redis") is None:
        redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
        log.info("notification.startup", redis_url=redis_url)
        r = aioredis.from_url(redis_url, decode_responses=False)
        app_state["redis"] = r
    if app_state.get("storage") is None:
        from .storage import NotificationStorage

        app_state["storage"] = NotificationStorage(app_state["redis"])


@app.on_event("shutdown")
async def shutdown() -> None:
    r = app_state.get("redis")
    if r is not None and hasattr(r, "aclose"):
        await r.aclose()
    log.info("notification.shutdown")


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "notification"}
