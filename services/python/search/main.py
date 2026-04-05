from __future__ import annotations

from pathlib import Path

import structlog
from fastapi import FastAPI

from .index import SearchIndex
from .routes import router

structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ],
)
log = structlog.get_logger()

app = FastAPI(title="ZONDEX Search Service", version="1.0.0")
app.include_router(router)

app_state: dict = {}


def _catalog_path() -> Path:
    return Path(__file__).resolve().parent / "data" / "catalog.json"


@app.on_event("startup")
async def startup() -> None:
    path = _catalog_path()
    log.info("search.startup", catalog_path=str(path))
    app_state["index"] = SearchIndex.load_from_file(path)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "search"}
