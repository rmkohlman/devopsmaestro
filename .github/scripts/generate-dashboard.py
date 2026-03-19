#!/usr/bin/env python3
"""
generate-dashboard.py — Generate Mermaid dashboard markdown from GitHub Project JSON.

Usage:
    gh project item-list 1 --owner rmkohlman --format json --limit 200 \
        | python3 .github/scripts/generate-dashboard.py > dashboard.md

The script reads project JSON from stdin and writes full Mermaid dashboard
markdown to stdout. The workflow then pipes that output to `gh issue edit`.
"""

import json
import sys
from collections import Counter, defaultdict
from datetime import date, timedelta


# ---------------------------------------------------------------------------
# Data loading
# ---------------------------------------------------------------------------

def load_items() -> list[dict]:
    """Read project JSON from stdin and return the items list."""
    data = json.load(sys.stdin)
    return data.get("items", [])


# ---------------------------------------------------------------------------
# Data helpers
# ---------------------------------------------------------------------------

def _label_names(item: dict) -> list[str]:
    return [l.lower() for l in item.get("labels", [])]


def _priority(item: dict) -> str:
    for lbl in _label_names(item):
        if lbl.startswith("priority:"):
            return lbl.split(":", 1)[1].strip()
    return "none"


def _is_epic(item: dict) -> bool:
    return "epic" in _label_names(item)


def _issue_number(item: dict) -> int | None:
    content = item.get("content", {})
    return content.get("number")


def _status(item: dict) -> str:
    return item.get("status") or "None"


def _agent(item: dict) -> str:
    return item.get("agent") or "unassigned"


def _effort(item: dict) -> str:
    return item.get("effort") or "None"


def _sprint(item: dict) -> str | None:
    return item.get("sprint") or None


# ---------------------------------------------------------------------------
# Chart 1 — Pie: Status distribution
# ---------------------------------------------------------------------------

def chart_status_pie(items: list[dict]) -> str:
    counts = Counter(_status(i) for i in items)
    total = len(items)
    lines = [
        f'```mermaid',
        f'pie title Project Status ({total} items)',
    ]
    # Preferred order
    for label in ["Done", "Todo", "In Progress"]:
        if label in counts:
            lines.append(f'    "{label}" : {counts[label]}')
    # Any unexpected statuses
    for label, cnt in counts.items():
        if label not in ("Done", "Todo", "In Progress"):
            lines.append(f'    "{label}" : {cnt}')
    lines.append("```")
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Chart 2 — Pie: Agent workload
# ---------------------------------------------------------------------------

def chart_agent_pie(items: list[dict]) -> str:
    counts = Counter(_agent(i) for i in items)
    total = len(items)
    lines = [
        "```mermaid",
        f'pie title Work by Agent ({total} items)',
    ]
    for agent, cnt in counts.most_common():
        lines.append(f'    "{agent}" : {cnt}')
    lines.append("```")
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Chart 3 — Pie: Effort breakdown
# ---------------------------------------------------------------------------

EFFORT_LABELS = {"S": "Small (hours)", "M": "Medium (1-2 days)", "L": "Large (3+ days)"}


def chart_effort_pie(items: list[dict]) -> str:
    counts = Counter(_effort(i) for i in items)
    total = len(items)
    lines = [
        "```mermaid",
        f'pie title Effort Distribution ({total} items)',
    ]
    for key, label in EFFORT_LABELS.items():
        if key in counts:
            lines.append(f'    "{label}" : {counts[key]}')
    if "None" in counts:
        lines.append(f'    "Unestimated" : {counts["None"]}')
    lines.append("```")
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Chart 4 — Bar (xychart-beta): Sprint velocity
# ---------------------------------------------------------------------------

def chart_sprint_velocity(items: list[dict]) -> str:
    """Completed items per sprint (Done status only)."""
    done_by_sprint: dict[str, int] = defaultdict(int)
    hotfixes = 0

    for item in items:
        if _status(item) != "Done":
            continue
        sprint = _sprint(item)
        if sprint:
            done_by_sprint[sprint] += 1
        else:
            hotfixes += 1

    # Sort sprints naturally
    sprint_names = sorted(done_by_sprint.keys())
    labels = sprint_names + (["Hotfixes (no sprint)"] if hotfixes else [])
    values = [done_by_sprint[s] for s in sprint_names] + ([hotfixes] if hotfixes else [])

    if not values:
        return "<!-- No completed sprint data -->"

    max_val = max(values) if values else 10
    y_max = max_val + 5

    label_str = ", ".join(f'"{l}"' for l in labels)
    value_str = ", ".join(str(v) for v in values)

    lines = [
        "```mermaid",
        "xychart-beta",
        '    title "Items Completed per Sprint"',
        f'    x-axis [{label_str}]',
        f'    y-axis "Items Completed" 0 --> {y_max}',
        f'    bar [{value_str}]',
        "```",
    ]
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Chart 5 — Gantt: Current sprint timeline
# ---------------------------------------------------------------------------

def chart_gantt(items: list[dict]) -> str:
    """Gantt chart for the most recent sprint's items."""
    # Find the latest sprint
    sprint_names = sorted(
        [s for s in {_sprint(i) for i in items} if s is not None],
        reverse=True,
    )
    if not sprint_names:
        return "<!-- No sprint data available -->"

    current_sprint = sprint_names[0]
    sprint_items = [i for i in items if _sprint(i) == current_sprint]

    today = date.today()
    # Bucket items by status for gantt sections
    done_items = [i for i in sprint_items if _status(i) == "Done"]
    active_items = [i for i in sprint_items if _status(i) == "In Progress"]
    todo_items = [i for i in sprint_items if _status(i) == "Todo"]

    lines = [
        "```mermaid",
        "gantt",
        f'    title {current_sprint} Timeline',
        "    dateFormat YYYY-MM-DD",
        "    axisFormat %b %d",
        "    todayMarker stroke-width:3px,stroke:#f00",
    ]

    # Spread Done items across last ~2 weeks (cosmetic, we have no real dates)
    if done_items:
        lines.append("")
        lines.append("    section Completed")
        offset = len(done_items)
        for idx, item in enumerate(done_items):
            num = _issue_number(item) or "?"
            title = item.get("title", "")[:40].replace(":", "").replace(",", "")
            start = today - timedelta(days=offset - idx + 2)
            safe_id = f"t{num}"
            lines.append(f'    {title} #{num}  :done, {safe_id}, {start.strftime("%Y-%m-%d")}, 1d')

    if active_items:
        lines.append("")
        lines.append("    section In Progress")
        for item in active_items:
            num = _issue_number(item) or "?"
            title = item.get("title", "")[:40].replace(":", "").replace(",", "")
            safe_id = f"t{num}"
            lines.append(f'    {title} #{num}  :active, {safe_id}, {today.strftime("%Y-%m-%d")}, 3d')

    if todo_items:
        lines.append("")
        lines.append("    section Todo")
        for item in todo_items:
            num = _issue_number(item) or "?"
            title = item.get("title", "")[:40].replace(":", "").replace(",", "")
            safe_id = f"t{num}"
            lines.append(f'    {title} #{num}  :{safe_id}, {today.strftime("%Y-%m-%d")}, 2d')

    lines.append("```")
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Chart 6 — Flowchart: Active epics
# ---------------------------------------------------------------------------

def chart_epic_flowchart(items: list[dict]) -> str:
    """
    For each open epic, build a flowchart showing child issues.
    Children are discovered heuristically: issues in the same sprint
    or with related labels (not epics themselves) grouped under the epic.
    Since the project JSON doesn't encode parent/child relationships,
    we detect children by title patterns and co-membership in sprint buckets.
    """
    open_epics = [i for i in items if _is_epic(i) and _status(i) != "Done"]
    if not open_epics:
        return "<!-- No active epics -->"

    # Build a map: issue_number -> item
    by_number: dict[int, dict] = {}
    for item in items:
        num = _issue_number(item)
        if num is not None:
            by_number[num] = item

    # For each epic, find items that share the same sprint (excluding other epics)
    blocks = []
    for epic in open_epics:
        epic_num = _issue_number(epic)
        epic_title = epic.get("title", f"Epic #{epic_num}")
        epic_sprint = _sprint(epic)

        if epic_sprint:
            children = [
                i for i in items
                if _sprint(i) == epic_sprint
                and not _is_epic(i)
                and _issue_number(i) != epic_num
            ]
        else:
            children = []

        if not children:
            continue

        done_children = [i for i in children if _status(i) == "Done"]
        active_children = [i for i in children if _status(i) == "In Progress"]
        todo_children = [i for i in children if _status(i) == "Todo"]

        lines = [
            "```mermaid",
            "flowchart LR",
        ]

        if done_children:
            lines.append('    subgraph "✅ Complete"')
            prev = None
            for child in done_children:
                num = _issue_number(child)
                short = child.get("title", "")[:25].replace('"', "'")
                node = f'N{num}["#{num} {short}"]'
                if prev:
                    lines.append(f"        {prev} --> {node}")
                else:
                    lines.append(f"        {node}")
                prev = f"N{num}"
            lines.append("    end")

        if active_children:
            lines.append('    subgraph "🔄 In Progress"')
            for child in active_children:
                num = _issue_number(child)
                short = child.get("title", "")[:25].replace('"', "'")
                lines.append(f'        N{num}["#{num} {short}"]')
            lines.append("    end")

        if todo_children:
            lines.append('    subgraph "🔲 Remaining"')
            for child in todo_children:
                num = _issue_number(child)
                short = child.get("title", "")[:25].replace('"', "'")
                lines.append(f'        N{num}["#{num} {short}"]')
            lines.append("    end")

        # Style nodes by status
        for child in done_children:
            lines.append(f'    style N{_issue_number(child)} fill:#2da44e,color:#fff')
        for child in active_children:
            lines.append(f'    style N{_issue_number(child)} fill:#bf8700,color:#fff')
        for child in todo_children:
            lines.append(f'    style N{_issue_number(child)} fill:#666,color:#fff')

        lines.append("```")
        blocks.append((epic_title, "\n".join(lines)))

    if not blocks:
        return "<!-- No active epics with sprint children -->"

    sections = []
    for epic_title, chart in blocks:
        sections.append(f"#### {epic_title}\n\n{chart}")
    return "\n\n".join(sections)


# ---------------------------------------------------------------------------
# Chart 7 — Bar (xychart-beta): Agent completion rates
# ---------------------------------------------------------------------------

def chart_agent_completion(items: list[dict]) -> str:
    done_by_agent: dict[str, int] = defaultdict(int)
    remaining_by_agent: dict[str, int] = defaultdict(int)

    for item in items:
        agent = _agent(item)
        if agent == "unassigned":
            continue
        if _status(item) == "Done":
            done_by_agent[agent] += 1
        else:
            remaining_by_agent[agent] += 1

    # All agents that have any work (done or remaining)
    all_agents = sorted(
        {a for a in list(done_by_agent) + list(remaining_by_agent)},
        key=lambda a: -(done_by_agent.get(a, 0) + remaining_by_agent.get(a, 0)),
    )

    if not all_agents:
        return "<!-- No agent data -->"

    done_vals = [done_by_agent.get(a, 0) for a in all_agents]
    remaining_vals = [remaining_by_agent.get(a, 0) for a in all_agents]
    y_max = max(max(done_vals), max(remaining_vals)) + 5

    label_str = ", ".join(f'"{a}"' for a in all_agents)
    done_str = ", ".join(str(v) for v in done_vals)
    remaining_str = ", ".join(str(v) for v in remaining_vals)

    lines = [
        "```mermaid",
        "xychart-beta",
        '    title "Agent Work: Done vs Remaining"',
        f'    x-axis [{label_str}]',
        f'    y-axis "Items" 0 --> {y_max}',
        f'    bar [{done_str}]',
        f'    bar [{remaining_str}]',
        "```",
    ]
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Chart 8 — Table: Backlog by priority
# ---------------------------------------------------------------------------

def chart_backlog_table(items: list[dict]) -> str:
    """Markdown table of open items organized by priority label."""
    priority_order = ["high", "medium", "low", "none"]
    priority_emoji = {"high": "🔴 **High**", "medium": "🟡 **Medium**", "low": "🟢 **Low**", "none": "⚪ **None**"}

    open_items = [i for i in items if _status(i) != "Done"]

    by_priority: dict[str, list[dict]] = defaultdict(list)
    for item in open_items:
        by_priority[_priority(item)].append(item)

    rows = []
    for prio in priority_order:
        prio_items = by_priority.get(prio, [])
        if not prio_items:
            continue
        # Group by agent within priority
        by_agent: dict[str, list[dict]] = defaultdict(list)
        for item in prio_items:
            by_agent[_agent(item)].append(item)

        for agent, agent_items in sorted(by_agent.items()):
            refs = ", ".join(
                f'#{_issue_number(i)}' for i in sorted(agent_items, key=lambda x: _issue_number(x) or 0)
            )
            # Truncate very long ref lists
            if len(refs) > 80:
                refs = refs[:77] + "..."
            rows.append(f"| {priority_emoji[prio]} | {agent} | {refs} |")

    if not rows:
        return "<!-- No open backlog items -->"

    lines = [
        "| Priority | Agent | Issues |",
        "|----------|-------|--------|",
    ] + rows
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Full dashboard assembly
# ---------------------------------------------------------------------------

def generate_dashboard(items: list[dict]) -> str:
    today_str = date.today().strftime("%Y-%m-%d")

    # Find latest sprint name for section heading
    sprint_names = sorted(
        [s for s in {_sprint(i) for i in items} if s is not None],
        reverse=True,
    )
    current_sprint = sprint_names[0] if sprint_names else "Current Sprint"

    sections = [
        f"## 📊 DevOpsMaestro Project Dashboard",
        "",
        f"> **Last updated:** {today_str} | **Auto-generated from project data**",
        "",
        "This issue contains visual dashboards rendered with Mermaid charts. "
        "Pin this issue for at-a-glance project visibility.",
        "",
        "---",
        "",
        "### Project Status Distribution",
        "",
        chart_status_pie(items),
        "",
        "---",
        "",
        "### Agent Workload Distribution",
        "",
        chart_agent_pie(items),
        "",
        "---",
        "",
        "### Effort Breakdown",
        "",
        chart_effort_pie(items),
        "",
        "---",
        "",
        "### Sprint Velocity",
        "",
        chart_sprint_velocity(items),
        "",
        "---",
        "",
        f"### {current_sprint} Progress",
        "",
        chart_gantt(items),
        "",
        "---",
        "",
        "### Active Epic Flowcharts",
        "",
        chart_epic_flowchart(items),
        "",
        "---",
        "",
        "### Agent Completion Rates",
        "",
        chart_agent_completion(items),
        "",
        "---",
        "",
        "### Backlog by Priority",
        "",
        chart_backlog_table(items),
        "",
        "---",
        "",
        "## How to Update This Dashboard",
        "",
        "This dashboard is auto-generated by the `update-dashboard` GitHub Actions workflow.",
        "To trigger a manual refresh:",
        "```bash",
        "gh workflow run update-dashboard.yml --repo rmkohlman/devopsmaestro",
        "```",
        "Or push to `main` / wait for the Monday 6 AM UTC schedule.",
    ]

    return "\n".join(sections)


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    items = load_items()
    if not items:
        print("ERROR: No items found in project JSON", file=sys.stderr)
        sys.exit(1)
    print(generate_dashboard(items))
