import logging
import os
import shutil


def copy(src, dst):
    """
    Copy a file or directory from src to dst, sync data to disk, and clear cache.
    """
    try:
        if os.path.isfile(src):
            dst_file = (
                os.path.join(dst, os.path.basename(src)) if os.path.isdir(dst) else dst
            )
            shutil.copy2(src, dst_file)
            _sync_and_clear_cache(dst_file)
        elif os.path.isdir(src):
            if not os.path.exists(dst):
                os.makedirs(dst)
            elif os.path.isfile(dst):
                raise ValueError(
                    f"Destination {dst} is a file, cannot copy directory into a file."
                )
            for root, dirs, files in os.walk(src):
                relative_path = os.path.relpath(root, src)
                dest_dir = os.path.join(dst, relative_path)
                if not os.path.exists(dest_dir):
                    os.makedirs(dest_dir)
                for file in files:
                    src_file = os.path.join(root, file)
                    dst_file = os.path.join(dest_dir, file)
                    shutil.copy2(src_file, dst_file)
                    _sync_and_clear_cache(dst_file)
        else:
            logging.warning(f"Source {src} does not exist.")
    except Exception as e:
        logging.error(f"Error copying from {src} to {dst}: {e}")
        raise


def delete(path):
    """
    Delete the file or directory at path, ensuring data is synced and cache is cleared.
    """
    try:
        if os.path.isfile(path):
            _sync_and_clear_cache(path)
            os.remove(path)
        elif os.path.isdir(path):
            for root, dirs, files in os.walk(path, topdown=False):
                for file in files:
                    file_path = os.path.join(root, file)
                    _sync_and_clear_cache(file_path)
                    os.remove(file_path)
                for dir in dirs:
                    dir_path = os.path.join(root, dir)
                    os.rmdir(dir_path)
            os.rmdir(path)
        else:
            logging.warning(f"Path {path} does not exist.")
    except Exception as e:
        logging.error(f"Error deleting {path}: {e}")
        raise


def move(src, dst):
    """
    Move a file or directory from src to dst, sync data to disk, and clear cache.
    """
    try:
        if os.path.exists(src):
            if os.path.isfile(src):
                _sync_and_clear_cache(src)
                dst_file = (
                    os.path.join(dst, os.path.basename(src))
                    if os.path.isdir(dst)
                    else dst
                )
                shutil.move(src, dst_file)
                _sync_and_clear_cache(dst_file)
            elif os.path.isdir(src):
                if os.path.exists(dst) and os.path.isfile(dst):
                    raise ValueError(
                        f"Destination {dst} is a file, cannot move directory into a file."
                    )
                shutil.move(src, dst)
                for root, dirs, files in os.walk(dst):
                    for file in files:
                        file_path = os.path.join(root, file)
                        _sync_and_clear_cache(file_path)
            else:
                logging.warning(f"Source {src} is neither a file nor a directory.")
                return
        else:
            logging.warning(f"Source {src} does not exist.")
    except Exception as e:
        logging.error(f"Error moving from {src} to {dst}: {e}")
        raise


def clear_cache(path):
    """
    Clear cache for the file or directory at the given path.
    """
    try:
        if os.path.isfile(path):
            _sync_and_clear_cache(path)
        elif os.path.isdir(path):
            for root, dirs, files in os.walk(path):
                for file in files:
                    file_path = os.path.join(root, file)
                    _sync_and_clear_cache(file_path)
        else:
            logging.warning(f"Path {path} does not exist.")
    except Exception as e:
        logging.error(f"Error clearing cache for {path}: {e}")
        raise


def _sync_and_clear_cache(file_path):
    """
    Sync data to disk and clear cache for a single file.
    """
    try:
        if os.path.isfile(file_path):
            with open(file_path, "rb") as f:
                f.flush()
                os.fsync(f.fileno())
                _advise_drop_cache(f.fileno())
        else:
            logging.warning(f"Cannot sync and clear cache for {file_path}: Not a file.")
    except Exception as e:
        logging.error(f"Error syncing and clearing cache for {file_path}: {e}")
        raise


def _advise_drop_cache(fileno):
    """
    Advise the kernel to drop cache for the file given by file descriptor.
    """
    # Note: posix_fadvise is not supported for mac M1, hence the warning. For Mac, fsync seems to be enough
    try:
        if hasattr(os, "posix_fadvise") and hasattr(os, "POSIX_FADV_DONTNEED"):
            os.posix_fadvise(fileno, 0, 0, os.POSIX_FADV_DONTNEED)
        else:
            logging.warning(
                "posix_fadvise not available on this system. Cannot clear cache for file."
            )
    except Exception as e:
        logging.error(f"Error in posix_fadvise: {e}")
        raise
