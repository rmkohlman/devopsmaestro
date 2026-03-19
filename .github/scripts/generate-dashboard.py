#!/usr/bin/env python3
"""
generate-dashboard.py — Generate dashboard markdown from GitHub Project JSON.

Usage (Mermaid issue dashboard — default):
    gh project item-list 1 --owner rmkohlman --format json --limit 200 \
        | python3 .github/scripts/generate-dashboard.py > dashboard.md

Usage (Project README — no Mermaid):
    gh project item-list 1 --owner rmkohlman --format json --limit 200 \
        | python3 .github/scripts/generate-dashboard.py --readme > readme.md

The script reads project JSON from stdin and writes markdown to stdout.
Default mode: Mermaid charts for issue #118.
--readme mode:  Plain markdown tables/text for the GitHub Project README tab.
"""

import argparse
import json
import sys
from collections import Counter, defaultdict
from datetime import date, timedelta, timezone


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
# README mode — plain markdown, NO Mermaid
# ---------------------------------------------------------------------------

def readme_health_table(items: list[dict]) -> str:
    """Health stats summary table."""
    total = len(items)
    done = sum(1 for i in items if _status(i) == "Done")
    in_progress = sum(1 for i in items if _status(i) == "In Progress")
    todo = sum(1 for i in items if _status(i) == "Todo")
    pct = int(done / total * 100) if total else 0

    lines = [
        "| Metric | Value |",
        "|--------|-------|",
        f"| Total items | {total} |",
        f"| ✅ Done | {done} |",
        f"| 🔄 In Progress | {in_progress} |",
        f"| 🔲 Todo | {todo} |",
        f"| Completion | {pct}% |",
    ]
    return "\n".join(lines)


def readme_active_epics(items: list[dict]) -> str:
    """
    Find items whose title contains 'epic' (case-insensitive) that are In Progress,
    and list them with a child-count summary using the same heuristic as the Mermaid chart.
    """
    active_epics = [
        i for i in items
        if "epic" in i.get("title", "").lower() and _status(i) == "In Progress"
    ]

    if not active_epics:
        return "_No active epics._"

    rows = ["| Epic | Sprint | Done | Todo | In Progress |",
            "|------|--------|------|------|-------------|"]

    for epic in active_epics:
        epic_num = _issue_number(epic)
        epic_title = epic.get("title", f"Epic #{epic_num}")
        epic_sprint = _sprint(epic)

        if epic_sprint:
            children = [
                i for i in items
                if _sprint(i) == epic_sprint
                and "epic" not in i.get("title", "").lower()
                and _issue_number(i) != epic_num
            ]
        else:
            children = []

        done_c = sum(1 for c in children if _status(c) == "Done")
        ip_c = sum(1 for c in children if _status(c) == "In Progress")
        todo_c = sum(1 for c in children if _status(c) == "Todo")
        sprint_str = epic_sprint or "—"
        rows.append(f"| {epic_title} | {sprint_str} | {done_c} | {todo_c} | {ip_c} |")

    return "\n".join(rows)


def readme_current_sprint(items: list[dict]) -> str:
    """
    Find the sprint with the most non-Done items and list them with their status.
    """
    # Count non-done items per sprint
    active_by_sprint: dict[str, int] = defaultdict(int)
    for item in items:
        s = _sprint(item)
        if s and _status(item) != "Done":
            active_by_sprint[s] += 1

    if not active_by_sprint:
        return "_No active sprint items._"

    current_sprint = max(active_by_sprint, key=lambda s: active_by_sprint[s])
    sprint_items = [i for i in items if _sprint(i) == current_sprint]

    # Sort: In Progress first, then Todo, then Done
    order = {"In Progress": 0, "Todo": 1, "Done": 2}
    sprint_items.sort(key=lambda i: (order.get(_status(i), 9), _issue_number(i) or 0))

    status_emoji = {"In Progress": "🔄", "Todo": "🔲", "Done": "✅"}

    lines = [f"**Sprint: {current_sprint}** ({active_by_sprint[current_sprint]} open items)\n"]
    lines.append("| # | Title | Status | Agent |")
    lines.append("|---|-------|--------|-------|")
    for item in sprint_items:
        num = _issue_number(item) or "—"
        title = item.get("title", "")[:60]
        st = _status(item)
        emoji = status_emoji.get(st, "")
        agent = _agent(item)
        lines.append(f"| {num} | {title} | {emoji} {st} | {agent} |")

    return "\n".join(lines)


def readme_agent_workload(items: list[dict]) -> str:
    """Agent workload table: done / in-progress / todo / total."""
    stats: dict[str, dict[str, int]] = defaultdict(lambda: {"Done": 0, "In Progress": 0, "Todo": 0})
    for item in items:
        a = _agent(item)
        s = _status(item)
        if s in stats[a]:
            stats[a][s] += 1
        else:
            stats[a]["Todo"] += 0  # ensure key exists

    # Sort by total descending
    sorted_agents = sorted(stats.items(), key=lambda kv: sum(kv[1].values()), reverse=True)

    lines = [
        "| Agent | ✅ Done | 🔄 In Progress | 🔲 Todo | Total |",
        "|-------|--------|----------------|--------|-------|",
    ]
    for agent, counts in sorted_agents:
        d = counts.get("Done", 0)
        ip = counts.get("In Progress", 0)
        td = counts.get("Todo", 0)
        total = d + ip + td
        lines.append(f"| {agent} | {d} | {ip} | {td} | {total} |")

    return "\n".join(lines)


def readme_sprint_velocity(items: list[dict]) -> str:
    """Sprint velocity table: completed items per sprint."""
    done_by_sprint: dict[str, int] = defaultdict(int)
    hotfixes = 0

    for item in items:
        if _status(item) != "Done":
            continue
        s = _sprint(item)
        if s:
            done_by_sprint[s] += 1
        else:
            hotfixes += 1

    if not done_by_sprint and not hotfixes:
        return "_No completed sprint data._"

    sprint_names = sorted(done_by_sprint.keys())
    lines = [
        "| Sprint | Items Completed |",
        "|--------|----------------|",
    ]
    for s in sprint_names:
        lines.append(f"| {s} | {done_by_sprint[s]} |")
    if hotfixes:
        lines.append(f"| _(no sprint / hotfixes)_ | {hotfixes} |")

    return "\n".join(lines)


def readme_effort_distribution(items: list[dict]) -> str:
    """Effort distribution table."""
    counts = Counter(_effort(i) for i in items)
    total = len(items)

    effort_map = [
        ("S", "Small (hours)"),
        ("M", "Medium (1-2 days)"),
        ("L", "Large (3+ days)"),
        ("None", "Unestimated"),
    ]

    lines = [
        "| Effort | Label | Count | % |",
        "|--------|-------|-------|---|",
    ]
    for key, label in effort_map:
        cnt = counts.get(key, 0)
        if cnt == 0:
            continue
        pct = int(cnt / total * 100) if total else 0
        lines.append(f"| {key} | {label} | {cnt} | {pct}% |")

    return "\n".join(lines)


def readme_backlog_summary(items: list[dict]) -> str:
    """
    Count of Todo items NOT in any sprint, grouped by agent.
    """
    backlog = [i for i in items if _status(i) == "Todo" and _sprint(i) is None]

    if not backlog:
        return "_Backlog is empty — all Todo items are assigned to a sprint._"

    by_agent: dict[str, int] = defaultdict(int)
    for item in backlog:
        by_agent[_agent(item)] += 1

    lines = [
        f"**{len(backlog)} unscheduled Todo items** (not in any sprint)\n",
        "| Agent | Unscheduled Items |",
        "|-------|------------------|",
    ]
    for agent, cnt in sorted(by_agent.items(), key=lambda kv: -kv[1]):
        lines.append(f"| {agent} | {cnt} |")

    return "\n".join(lines)


def generate_readme(items: list[dict]) -> str:
    """
    Generate a Project README version of the dashboard.
    No Mermaid — tables, bold, links, and emoji only.
    """
    now_utc = date.today().strftime("%Y-%m-%d")

    sections = [
        "# DevOpsMaestro Toolkit — Project Dashboard",
        "",
        "> Auto-generated from live GitHub Project data. **Do not edit manually.**",
        "",
        "---",
        "",
        "## Project Health",
        "",
        readme_health_table(items),
        "",
        "---",
        "",
        "## Active Epics",
        "",
        readme_active_epics(items),
        "",
        "---",
        "",
        "## Current Sprint",
        "",
        readme_current_sprint(items),
        "",
        "---",
        "",
        "## Agent Workload",
        "",
        readme_agent_workload(items),
        "",
        "---",
        "",
        "## Sprint Velocity",
        "",
        readme_sprint_velocity(items),
        "",
        "---",
        "",
        "## Effort Distribution",
        "",
        readme_effort_distribution(items),
        "",
        "---",
        "",
        "## Backlog Summary",
        "",
        readme_backlog_summary(items),
        "",
        "---",
        "",
        "## Quick Links",
        "",
        "| Resource | Link |",
        "|----------|------|",
        "| Main repo | [rmkohlman/devopsmaestro](https://github.com/rmkohlman/devopsmaestro) |",
        "| Homebrew tap | [rmkohlman/homebrew-tap](https://github.com/rmkohlman/homebrew-tap) |",
        "| Docs | [devopsmaestro.dev](https://devopsmaestro.dev) |",
        "| Dashboard issue | [#118](https://github.com/rmkohlman/devopsmaestro/issues/118) |",
        "| Releases | [Releases](https://github.com/rmkohlman/devopsmaestro/releases) |",
        "",
        "---",
        "",
        "## Field Guide",
        "",
        "| Field | Values |",
        "|-------|--------|",
        "| **Status** | Todo · In Progress · Done |",
        "| **Agent** | release · dvm-core · test · document · nvim · theme · terminal · sdk · database · architecture · cli-architect · security |",
        "| **Sprint** | Sprint N (sequential) |",
        "| **Effort** | S (hours) · M (1-2 days) · L (3+ days) |",
        "| **Priority** | high · medium · low |",
        "",
        "---",
        "",
        "## Recommended Views",
        "",
        "| View | Filter |",
        "|------|--------|",
        "| Current sprint | `sprint:\"Sprint N\" status:Todo,\"In Progress\"` |",
        "| My work | `agent:release status:\"In Progress\"` |",
        "| Open bugs | `label:\"type: bug\" status:Todo,\"In Progress\"` |",
        "| Unscheduled backlog | `no:sprint status:Todo` |",
        "| High priority | `label:\"priority: high\" status:Todo,\"In Progress\"` |",
        "",
        "---",
        "",
        "## How to Refresh",
        "",
        "This README is auto-generated by the `update-dashboard` GitHub Actions workflow.",
        "To trigger a manual refresh:",
        "",
        "```",
        "gh workflow run update-dashboard.yml --repo rmkohlman/devopsmaestro",
        "```",
        "",
        "---",
        "",
        f"_Last updated: {now_utc} UTC_",
    ]

    return "\n".join(sections)


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Generate dashboard markdown from GitHub Project JSON."
    )
    parser.add_argument(
        "--readme",
        action="store_true",
        help="Output Project README format (no Mermaid) instead of issue dashboard.",
    )
    args = parser.parse_args()

    items = load_items()
    if not items:
        print("ERROR: No items found in project JSON", file=sys.stderr)
        sys.exit(1)

    if args.readme:
        print(generate_readme(items))
    else:
        print(generate_dashboard(items))
