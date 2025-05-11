# CDX (Change Directory Xplorer) — Filesystem Explorer Concept

CDX is a terminal-based file explorer inspired by the interface seen in the series *Silo*.  
The TUI is structured into three distinct sections:

## 1. Top Bar

Displays:
- Breadcrumb-style navigation (e.g., `/home/user/projects`)
- Current sorting mode (e.g., `Sort: Date ↓`)

```

┌─────────────────────────────────────────────────────────────┐
│ /home/user/projects                 Sort: Date Modified ↓   │
└─────────────────────────────────────────────────────────────┘

```

## 2. Main Content Area (Gallery View)

Files and directories are displayed in framed blocks, each showing:
- An icon (📁 folder, 📄 file)
- Truncated name if too long: `start..end.ext`
- Creation date and size

**Example display:**

```

┌────────────────────────────────┐
│    my_really_cool_script.sh    │
│     2025-11-28 22:45   1.2MB   │
└────────────────────────────────┘

┌────────────────────────────────┐
│           start here           │
│     2023-01-03 09:10   320KB   │
└────────────────────────────────┘

```

## 3. Bottom Bar

Acts as a shell passthrough for basic commands.

```

> mkdir new\_folder
> cd projects (this will also update the current visual dir on cdx)
> cdx sort name

```