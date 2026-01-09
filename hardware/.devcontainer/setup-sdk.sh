#!/bin/bash
# Patch SDK Kconfig (lxdialog binary build fix)
LXD_SCRIPT="sdk/tools/kconfig/lxdialog/check-lxdialog.sh"
if [ -f "$LXD_SCRIPT" ]; then
    printf "#!/bin/sh\nif [ \"\$1\" = \"-check\" ]; then exit 0; fi\necho '-DCURSES_LOC=\"<ncurses.h>\" -DNCURSES_WIDECHAR=1'\necho '-lncurses'\n" > "$LXD_SCRIPT"
    chmod +x "$LXD_SCRIPT"
fi

# Silence git safety warnings
git config --global --add safe.directory /workspaces/hardware
git config --global --add safe.directory /workspaces/hardware/sdk
