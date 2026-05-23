# integration-tooling-skills

Embedded pack for MCP and external app/service integration design.

Use this pack when a task needs MCP server planning, external app integration shaping, connector boundaries, or tool permission design.

Install it through the harness:

```bash
./skill-harness install --packs=integration-tooling-skills --packs-only
```

Prerequisites depend on the selected integration. Do not assume credentials, OAuth scopes, production access, or network permissions are available until the task explicitly grants them.

Included skills:

- `mcp-server-planning`
- `app-integration-shaping`
