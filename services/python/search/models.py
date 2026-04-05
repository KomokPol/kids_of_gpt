from pydantic import BaseModel, Field


class SearchQuery(BaseModel):
    q: str
    category: str | None = None
    limit: int = Field(default=20, le=100)


class CatalogItem(BaseModel):
    item_id: str
    title: str
    description: str
    category: str  # "barygi", "kinoshka", "balanda"
    price: int = 0
    min_level: int = 0
    tags: list[str] = []


class SearchResult(BaseModel):
    items: list[CatalogItem] = []
    total: int = 0
    query: str = ""
