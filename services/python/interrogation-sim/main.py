from __future__ import annotations

import structlog
from fastapi import FastAPI

from engine import InterrogationEngine
from routes import router

structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ],
)
log = structlog.get_logger()

app = FastAPI(
    title="Sharaga Interrogation Sim",
    version="1.0.0",
    description="Симуляция допроса / антискам сценариев для режима Sharaga",
)
app.include_router(router)

app_state: dict = {}


@app.on_event("startup")
async def startup() -> None:
    app_state["engine"] = InterrogationEngine()
    log.info("interrogation_sim.startup")


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "interrogation-sim"}
