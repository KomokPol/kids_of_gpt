from __future__ import annotations

from contextlib import asynccontextmanager
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

app_state: dict = {}


def _catalog_path() -> Path:
    return Path(__file__).resolve().parent / "data" / "catalog.json"


@asynccontextmanager
async def lifespan(app_inst: FastAPI):
    path = _catalog_path()
    log.info("search.startup", catalog_path=str(path))
    if not path.exists():
        log.error("search.catalog_not_found", path=str(path))
        raise FileNotFoundError(f"Catalog not found: {path}")
    app_state["index"] = SearchIndex.load_from_file(path)
    log.info("search.index_loaded", items=len(app_state["index"].all_items()))
    yield
    log.info("search.shutdown")


app = FastAPI(title="ZONDEX Search Service", version="1.0.0", lifespan=lifespan)
app.include_router(router)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "search"}
