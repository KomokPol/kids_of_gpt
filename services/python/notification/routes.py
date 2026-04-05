from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, Query

from .main import app_state
from .models import NotificationIn, NotificationList, NotificationStored
from .storage import NotificationStorage

router = APIRouter()


def get_storage() -> NotificationStorage:
    return app_state["storage"]


@router.post(
    "/notifications",
    response_model=NotificationStored,
    status_code=201,
)
async def send_notification(
    body: NotificationIn,
    storage: NotificationStorage = Depends(get_storage),
) -> NotificationStored:
    return await storage.save(body)


@router.get("/notifications/{user_id}/unread-count")
async def unread_count(
    user_id: str,
    storage: NotificationStorage = Depends(get_storage),
) -> dict:
    count = await storage.count_unread(user_id)
    return {"user_id": user_id, "unread_count": count}


@router.get("/notifications/{user_id}", response_model=NotificationList)
async def list_notifications(
    user_id: str,
    limit: int = Query(default=50, ge=1, le=500),
    storage: NotificationStorage = Depends(get_storage),
) -> NotificationList:
    notifications = await storage.list(user_id, limit=limit)
    total = await storage.count_total(user_id)
    return NotificationList(
        user_id=user_id,
        notifications=notifications,
        total=total,
    )


@router.patch("/notifications/{user_id}/{notification_id}/read")
async def mark_notification_read(
    user_id: str,
    notification_id: str,
    storage: NotificationStorage = Depends(get_storage),
) -> dict:
    ok = await storage.mark_read(user_id, notification_id)
    if not ok:
        raise HTTPException(status_code=404, detail="notification not found")
    return {"ok": True}
