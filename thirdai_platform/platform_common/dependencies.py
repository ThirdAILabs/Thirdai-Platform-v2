from fastapi import HTTPException, status

from .utils import disk_usage


def is_on_low_disk(threshold: float = 0.8):
    disk_stats = disk_usage()

    disk_used = disk_stats["used"] / disk_stats["total"]
    if disk_used > threshold:
        space_needed = ((disk_used - threshold) * disk_stats["total"]) / (
            1024 * 1024
        )  # MB
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=f"Platform is at {disk_used * 100:.2f}% disk usage. Clear at least {space_needed:.2f} MB space.",
        )
