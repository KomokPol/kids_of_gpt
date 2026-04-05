from __future__ import annotations

from fastapi import APIRouter, HTTPException, Query

from .index import SearchIndex
from .models import CatalogItem, SearchResult

router = APIRouter()


def get_index() -> SearchIndex:
    from . import main as main_mod

    return main_mod.app_state["index"]


@router.get("/search", response_model=SearchResult)
def search(
    q: str = Query(default=""),
    category: str | None = Query(default=None),
    limit: int = Query(default=20, le=100),
    search_enabled: bool = Query(default=True),
) -> SearchResult:
    if not search_enabled:
        raise HTTPException(
            status_code=403,
            detail="Поиск заблокирован. Достигните уровня 3 для разблокировки",
        )
    idx = get_index()
    items, total = idx.search(query=q, category=category, limit=limit)
    return SearchResult(items=items, total=total, query=q)


@router.get("/catalog", response_model=list[CatalogItem])
def catalog() -> list[CatalogItem]:
    return get_index().all_items()


@router.get("/catalog/{item_id}", response_model=CatalogItem)
def catalog_item(item_id: str) -> CatalogItem:
    item = get_index().get(item_id)
    if item is None:
        raise HTTPException(status_code=404, detail="Товар не найден")
    return item
