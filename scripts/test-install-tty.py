"""Exercise curl-pipe-style setup in a real PTY, without keys or a node."""
import errno
import os
import pathlib
import pty
import select
import signal
import sys
import time

work = pathlib.Path(sys.argv[1])
for skip in (False, True):
    marker = work / "ui-started"
    marker.unlink(missing_ok=True)
    pid, fd = pty.fork()
    if pid == 0:
        os.environ["TEST_UI_MARKER"] = str(marker)
        os.environ["AZ_NO_UI"] = "1" if skip else "0"
        os.execl("/bin/sh", "sh", "-c", 'cat "$TEST_INSTALLER" | sh')
    output = bytearray()
    status = None
    deadline = time.monotonic() + 20
    try:
        while time.monotonic() < deadline:
            if select.select([fd], [], [], 0.1)[0]:
                try:
                    data = os.read(fd, 65536)
                except OSError as exc:
                    if exc.errno != errno.EIO:
                        raise
                    data = b""
                output.extend(data)
            done, status = os.waitpid(pid, os.WNOHANG)
            if done:
                break
        else:
            os.killpg(pid, signal.SIGKILL)
            os.waitpid(pid, 0)
            raise AssertionError("installer/menu did not exit")
    finally:
        os.close(fd)
    assert os.waitstatus_to_exitcode(status) == 0, output.decode(errors="replace")
    if skip:
        assert not marker.exists(), "AZ_NO_UI did not prevent menu launch"
    else:
        assert marker.read_text().strip() == "ui", "piped install did not launch UI on the terminal"
print("Piped installer launches the terminal menu; AZ_NO_UI skips it.")
