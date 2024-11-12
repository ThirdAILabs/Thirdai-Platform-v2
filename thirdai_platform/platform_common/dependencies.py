from fastapi import HTTPException, status

from .utils import disk_usage


def is_on_low_disk(threshold: float = 0.8):
    def func(size: int = 0):
        disk_stats = disk_usage()

        disk_use = (disk_stats["used"] + size) / disk_stats["total"]
        if disk_use > threshold:
            space_needed = ((disk_use - threshold) * disk_stats["total"]) / (
                1024 * 1024
            )  # MB
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Platform is at {disk_use * 100:.2f}% disk usage. Clear at least {space_needed:.2f} MB space.",
            )

    return func
