from __future__ import annotations

import json
import re
from pathlib import Path

from .models import CatalogItem

_WORD_RE = re.compile(r"[\w\-]+", re.UNICODE)

WEIGHT_TITLE = 3
WEIGHT_DESCRIPTION = 1
WEIGHT_TAGS = 2


def _words(text: str) -> list[str]:
    return [w.lower() for w in _WORD_RE.findall(text)]


class SearchIndex:
    def __init__(self, items: list[CatalogItem]) -> None:
        self._items = items
        self._by_id: dict[str, CatalogItem] = {i.item_id: i for i in items}

    @classmethod
    def load_from_file(cls, path: str | Path) -> SearchIndex:
        raw = Path(path).read_text(encoding="utf-8")
        data = json.loads(raw)
        items = [CatalogItem(**row) for row in data]
        return cls(items)

    def get(self, item_id: str) -> CatalogItem | None:
        return self._by_id.get(item_id)

    def all_items(self) -> list[CatalogItem]:
        return list(self._items)

    def _score_item(self, item: CatalogItem, query_tokens: list[str]) -> float:
        title_words = _words(item.title)
        desc_words = _words(item.description)
        tag_words: list[str] = []
        for tag in item.tags:
            tag_words.extend(_words(tag))

        score = 0.0
        for tok in query_tokens:
            score += title_words.count(tok) * WEIGHT_TITLE
            score += desc_words.count(tok) * WEIGHT_DESCRIPTION
            score += tag_words.count(tok) * WEIGHT_TAGS
        return score

    def search(
        self,
        query: str,
        category: str | None = None,
        limit: int = 20,
    ) -> tuple[list[CatalogItem], int]:
        q = (query or "").strip().lower()
        query_tokens = [t for t in q.split() if t]
        if not query_tokens:
            return [], 0

        pool = self._items
        if category is not None:
            pool = [i for i in pool if i.category == category]

        scored: list[tuple[float, CatalogItem]] = []
        for item in pool:
            s = self._score_item(item, query_tokens)
            if s > 0:
                scored.append((s, item))

        scored.sort(key=lambda x: (-x[0], x[1].item_id))
        top = [it for _, it in scored[:limit]]
        return top, len(scored)
