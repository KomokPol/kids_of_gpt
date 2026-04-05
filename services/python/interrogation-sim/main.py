from __future__ import annotations

from contextlib import asynccontextmanager

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

app_state: dict = {}


@asynccontextmanager
async def lifespan(app_inst: FastAPI):
    app_state["engine"] = InterrogationEngine()
    log.info("interrogation_sim.startup", scenarios=len(app_state["engine"]._by_scenario_id))
    yield
    log.info("interrogation_sim.shutdown")


app = FastAPI(
    title="Sharaga Interrogation Sim",
    version="1.0.0",
    lifespan=lifespan,
)
app.include_router(router)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "interrogation-sim"}
