package lock

import (
	"fmt"
)

func getLockAcquireScript(path, owner string) string {
script := fmt.Sprintf(`
			set -eu

			LOCK_DIR='%s'
			INFO="$LOCK_DIR/info"

			if mkdir "$LOCK_DIR" 2>/dev/null; then
				printf 'pid=%%s\n' "$$" > "$INFO"
				printf 'owner=%%s\n' '%s' >> "$INFO"
				printf 'started=%%s\n' "$(date +%%s)" >> "$INFO"
				exit 0
			fi

			# Lock exists. Try to determine its owner.
			if [ ! -f "$INFO" ]; then
				# A partially-created/corrupt lock is safe to recover.
				rm -rf "$LOCK_DIR"

				if mkdir "$LOCK_DIR" 2>/dev/null; then
					printf 'pid=%%s\n' "$$" > "$INFO"
					printf 'owner=%%s\n' '%s' >> "$INFO"
					printf 'started=%%s\n' "$(date +%%s)" >> "$INFO"
					exit 0
				fi

				exit 1
			fi

			PID=$(sed -n 's/^pid=//p' "$INFO" | head -n 1)

			if [ -z "$PID" ]; then
				rm -rf "$LOCK_DIR"

				if mkdir "$LOCK_DIR" 2>/dev/null; then
					printf 'pid=%%s\n' "$$" > "$INFO"
					printf 'owner=%%s\n' '%s' >> "$INFO"
					printf 'started=%%s\n' "$(date +%%s)" >> "$INFO"
					exit 0
				fi

				exit 1
			fi

			# kill -0 checks whether the process exists without killing it.
			if kill -0 "$PID" 2>/dev/null; then
				exit 2
			fi

			# Owner is gone, so the lock is stale.
			rm -rf "$LOCK_DIR"

			if mkdir "$LOCK_DIR" 2>/dev/null; then
				printf 'pid=%%s\n' "$$" > "$INFO"
				printf 'owner=%%s\n' '%s' >> "$INFO"
				printf 'started=%%s\n' "$(date +%%s)" >> "$INFO"
				exit 0
			fi

			exit 1
			`, path, owner, owner, owner, owner)


return script
}

func getLockReleaseScript(path string) string {
	script := fmt.Sprintf(`
		set -eu

		LOCK_DIR='%s'
		INFO="$LOCK_DIR/info"

		[ -d "$LOCK_DIR" ] || exit 0
		[ -f "$INFO" ] || exit 0

		PID=$(sed -n 's/^pid=//p' "$INFO" | head -n 1)

		[ -n "$PID" ] || exit 0

		# Only remove our own lock.
		if [ "$PID" = "$$" ]; then
			rm -rf "$LOCK_DIR"
		fi
	`, path) 

return script
}


func getLockOwnerScript(path string) string {
	script := fmt.Sprintf(`
		set -eu

		LOCK_DIR='%s'
		INFO="$LOCK_DIR/info"

		[ -f "$INFO" ] || exit 1

		PID=$(sed -n 's/^pid=//p' "$INFO" | head -n 1)

		[ -n "$PID" ] || exit 1

		kill -0 "$PID" 2>/dev/null
	`,path)

return script
}



