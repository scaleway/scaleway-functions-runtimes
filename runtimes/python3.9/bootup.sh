# Script should be in the same directory as the bootup
SCRIPT_DIR=$(dirname "$0")
gunicorn --bind "127.0.0.1:$SCW_UPSTREAM_PORT" --error-logfile "-" --access-logfile "-" --chdir $SCRIPT_DIR index:app