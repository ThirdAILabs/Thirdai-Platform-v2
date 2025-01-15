from typing import Optional

from fastapi import HTTPException, status

from .utils import disk_usage


def is_on_low_disk(threshold: float = 0.8, path: Optional[str] = None):
    def func(size: int = 0):
        disk_stats = disk_usage(path=path)

        disk_use = (disk_stats["used"] + size) / disk_stats["total"]
        if disk_use > threshold:
            space_needed = ((disk_use - threshold) * disk_stats["total"]) / (
                1024 * 1024
            )  # MB
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=f"Platform is at {(disk_stats['used'] / disk_stats['total']) * 100:.2f}% disk usage. Clear at least {space_needed:.2f} MB space.",
            )

    return func
