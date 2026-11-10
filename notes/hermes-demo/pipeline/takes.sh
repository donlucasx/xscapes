set -e
cd "$(dirname "$0")"
DEMO_SCAPE=vista DEMO_IDLE=22 DEMO_PROMPT="lets get busy. Run a multi agent showcase of xscapes" python3 rec.py rec/vista > rec/vista.log 2>&1
DEMO_SCAPE=shore DEMO_PROMPT="count the lines of Go in this repo, per top-level folder, with one shell command" python3 rec.py rec/shore > rec/shore.log 2>&1
rm -f work/demo.sh
DEMO_SCAPE=harbour DEMO_PROMPT="think it over first, then create an empty file demo.sh here and chmod 777 it" python3 rec.py rec/harbour > rec/harbour.log 2>&1
xscapes scape vista; cat ~/.config/xscapes/scape
