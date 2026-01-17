# CI/CD Agent Pipeline

Run the nocturnal "lazy" implementation prompt automatically in your CI/CD pipeline. This guide shows how to set up a pipeline that uses OpenCode to implement the active proposal and create a merge/pull request with the changes.

## Prerequisites

### 1. Nocturnal Specification Workspace

Your project must have a nocturnal spec workspace initialized:

```bash
nocturnal spec init
```

### 2. Active Proposal

You need an active proposal with tasks defined in `implementation.md`:

```bash
nocturnal spec proposal add my-feature
nocturnal spec proposal activate my-feature
```

### 3. OpenCode Zen API Key

The pipeline uses [OpenCode Zen](https://opencode.ai/docs/zen) for model access. Get your API key:

1. Sign in at [opencode.ai/auth](https://opencode.ai/auth)
2. Add billing details
3. Copy your API key

Store this as a CI/CD secret variable named `OPENCODE_ZEN_API_KEY`.

### 4. Repository Access Token

The pipeline needs permission to push branches and create merge/pull requests.

**GitLab:** Create a Project Access Token with `write_repository` and `api` scopes, store as `GITLAB_TOKEN`.

**GitHub:** Create a Personal Access Token (classic) with `repo` scope, store as `GH_TOKEN`.

---

## GitLab CI/CD

Add this to your project's `.gitlab-ci.yml`:

```yaml
variables:
  NOCTURNAL_IMAGE: registry.gitlab.com/caffeinatedjack/nocturnal:latest

stages:
  - agent

agent:implement:
  stage: agent
  image: ${NOCTURNAL_IMAGE}
  rules:
    - if: $CI_PIPELINE_SOURCE == "web"      # Manual trigger from UI
    - if: $CI_PIPELINE_SOURCE == "trigger"  # API trigger
    - if: $CI_PIPELINE_SOURCE == "schedule" # Scheduled pipeline
  variables:
    GIT_STRATEGY: clone
    GIT_DEPTH: 0
  before_script:
    # Configure git
    - git config --global user.email "opencode@ci.local"
    - git config --global user.name "OpenCode CI"
    - git remote set-url origin "https://oauth2:${GITLAB_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.git"
    
    # Configure OpenCode authentication
    - mkdir -p ~/.local/share/opencode
    - |
      cat > ~/.local/share/opencode/auth.json << EOF
      {
        "opencode": {
          "type": "api",
          "key": "${OPENCODE_ZEN_API_KEY}"
        }
      }
      EOF
    
    # Configure OpenCode to use nocturnal MCP server
    - mkdir -p .opencode
    - |
      cat > .opencode/config.json << EOF
      {
        "mcp": {
          "nocturnal": {
            "type": "local",
            "enabled": true,
            "command": ["nocturnal", "mcp"]
          }
        }
      }
      EOF
  script:
    # Get the lazy prompt from nocturnal MCP
    - |
      PROMPT=$(cat << 'PROMPT_EOF'
      Use the nocturnal MCP server to implement the active proposal.

      1. Call the "context" tool to get the specification and design
      2. Call the "tasks" tool to get the current phase tasks
      3. Implement tasks using the "lazy" approach:
         - Implement quickly without extensive planning
         - If stuck, add TODO comments and move on
         - Mark tasks complete with "task_complete" tool
         - Continue until all tasks in the current phase are done
      4. Run tests at the end

      Be autonomous - don't ask for user input. Document any incomplete items.
      PROMPT_EOF
      )
    
    # Create feature branch
    - BRANCH_NAME="opencode/implement-$(date +%Y%m%d-%H%M%S)"
    - git checkout -b "${BRANCH_NAME}"
    
    # Run OpenCode with the lazy implementation prompt
    - opencode run --model opencode/claude-sonnet-4 "${PROMPT}"
    
    # Check for changes and create MR
    - |
      if [ -n "$(git status --porcelain)" ]; then
        git add -A
        git commit -m "feat: implement proposal tasks via OpenCode CI

      Automated implementation using nocturnal lazy prompt.
      
      Co-authored-by: OpenCode <opencode@ci.local>"
        
        git push -u origin "${BRANCH_NAME}"
        
        # Create merge request using GitLab API
        curl --silent --fail --request POST \
          --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
          --header "Content-Type: application/json" \
          --data "{
            \"source_branch\": \"${BRANCH_NAME}\",
            \"target_branch\": \"${CI_DEFAULT_BRANCH}\",
            \"title\": \"[OpenCode] Implement proposal tasks\",
            \"description\": \"Automated implementation created by OpenCode CI pipeline.\n\nReview the changes carefully before merging.\",
            \"remove_source_branch\": true
          }" \
          "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests"
        
        echo "Merge request created successfully"
      else
        echo "No changes were made"
      fi
```

### Triggering the Pipeline

**Manual trigger via UI:**
1. Go to CI/CD → Pipelines
2. Click "Run pipeline"
3. Select your branch

**Scheduled trigger:**
1. Go to CI/CD → Schedules
2. Create a new schedule (e.g., daily at 9am)

**API trigger:**
```bash
curl --request POST \
  --form "token=${TRIGGER_TOKEN}" \
  --form "ref=main" \
  "https://gitlab.com/api/v4/projects/${PROJECT_ID}/trigger/pipeline"
```

---

## GitHub Actions

Create `.github/workflows/opencode-agent.yml`:

```yaml
name: OpenCode Agent

on:
  workflow_dispatch:  # Manual trigger
  schedule:
    - cron: '0 9 * * 1-5'  # Weekdays at 9am UTC (optional)

env:
  NOCTURNAL_IMAGE: registry.gitlab.com/caffeinatedjack/nocturnal:latest

jobs:
  implement:
    runs-on: ubuntu-latest
    container:
      image: ${{ env.NOCTURNAL_IMAGE }}
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          token: ${{ secrets.GH_TOKEN }}

      - name: Configure git
        run: |
          git config --global user.email "opencode@ci.local"
          git config --global user.name "OpenCode CI"
          git config --global --add safe.directory "$GITHUB_WORKSPACE"

      - name: Configure OpenCode
        run: |
          # Authentication
          mkdir -p ~/.local/share/opencode
          cat > ~/.local/share/opencode/auth.json << EOF
          {
            "opencode": {
              "type": "api",
              "key": "${{ secrets.OPENCODE_ZEN_API_KEY }}"
            }
          }
          EOF
          
          # MCP server config
          mkdir -p .opencode
          cat > .opencode/config.json << EOF
          {
            "mcp": {
              "nocturnal": {
                "type": "local",
                "enabled": true,
                "command": ["nocturnal", "mcp"]
              }
            }
          }
          EOF

      - name: Create feature branch
        id: branch
        run: |
          BRANCH_NAME="opencode/implement-$(date +%Y%m%d-%H%M%S)"
          git checkout -b "${BRANCH_NAME}"
          echo "branch=${BRANCH_NAME}" >> $GITHUB_OUTPUT

      - name: Run OpenCode implementation
        run: |
          PROMPT=$(cat << 'PROMPT_EOF'
          Use the nocturnal MCP server to implement the active proposal.

          1. Call the "context" tool to get the specification and design
          2. Call the "tasks" tool to get the current phase tasks
          3. Implement tasks using the "lazy" approach:
             - Implement quickly without extensive planning
             - If stuck, add TODO comments and move on
             - Mark tasks complete with "task_complete" tool
             - Continue until all tasks in the current phase are done
          4. Run tests at the end

          Be autonomous - don't ask for user input. Document any incomplete items.
          PROMPT_EOF
          )
          
          opencode run --model opencode/claude-sonnet-4 "${PROMPT}"

      - name: Commit and push changes
        id: commit
        run: |
          if [ -n "$(git status --porcelain)" ]; then
            git add -A
            git commit -m "feat: implement proposal tasks via OpenCode CI

          Automated implementation using nocturnal lazy prompt.

          Co-authored-by: OpenCode <opencode@ci.local>"
            
            git push -u origin "${{ steps.branch.outputs.branch }}"
            echo "changes=true" >> $GITHUB_OUTPUT
          else
            echo "No changes were made"
            echo "changes=false" >> $GITHUB_OUTPUT
          fi

      - name: Create Pull Request
        if: steps.commit.outputs.changes == 'true'
        uses: actions/github-script@v7
        with:
          github-token: ${{ secrets.GH_TOKEN }}
          script: |
            const { data: pr } = await github.rest.pulls.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: '[OpenCode] Implement proposal tasks',
              head: '${{ steps.branch.outputs.branch }}',
              base: '${{ github.event.repository.default_branch }}',
              body: `Automated implementation created by OpenCode CI pipeline.
              
              Review the changes carefully before merging.
              
              ---
              *Generated by [nocturnal](https://gitlab.com/caffeinatedjack/nocturnal) + [OpenCode](https://opencode.ai)*`
            });
            
            console.log(`Pull request created: ${pr.html_url}`);
```

### Triggering the Workflow

**Manual trigger:**
1. Go to Actions tab
2. Select "OpenCode Agent" workflow
3. Click "Run workflow"

**Scheduled:** Runs automatically based on the cron schedule (remove if not needed).

---

## Environment Variables Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `OPENCODE_ZEN_API_KEY` | Yes | OpenCode Zen API key for model access |
| `GITLAB_TOKEN` | GitLab only | Project access token for pushing and creating MRs |
| `GH_TOKEN` | GitHub only | Personal access token for pushing and creating PRs |

---

## Model Selection

The examples use `opencode/claude-sonnet-4` via OpenCode Zen. Available models include:

| Model | Use Case |
|-------|----------|
| `opencode/claude-sonnet-4` | Recommended for most tasks |
| `opencode/claude-sonnet-4-5` | More capable, higher cost |
| `opencode/gpt-5.1-codex` | Alternative option |
| `opencode/claude-haiku-4-5` | Faster, lower cost |

See [OpenCode Zen docs](https://opencode.ai/docs/zen) for full model list and pricing.

---

## Customization

### Custom Goal

Pass a specific goal to focus the implementation:

```bash
opencode run --model opencode/claude-sonnet-4 "Implement the active proposal. Focus on: ${CUSTOM_GOAL}"
```

### Different Prompt Strategy

For more careful implementation, use the `start-implementation` prompt approach instead of `lazy`:

```bash
PROMPT="Use nocturnal MCP. Call context and tasks tools. Follow the start-implementation approach with investigation, test planning, implementation, and validation phases. Stop if you encounter blocking issues."
```

### Limiting Scope

To implement only specific tasks or phases, modify the prompt:

```bash
PROMPT="Use nocturnal MCP. Only implement tasks 1.1 and 1.2 from the current phase. Skip other tasks."
```

---

## Troubleshooting

### Pipeline fails with authentication error

Ensure `OPENCODE_ZEN_API_KEY` is set correctly and has valid credits.

### No changes made

- Check that an active proposal exists: `nocturnal spec proposal current`
- Verify tasks are defined in `spec/proposal/<name>/implementation.md`
- Check OpenCode logs for errors

### MCP server not found

Ensure the nocturnal binary is in PATH. The Docker image includes it at `/usr/local/bin/nocturnal`.

### Merge request creation fails

- Verify the access token has correct permissions
- Check that the target branch exists
- Ensure branch protection rules allow the token to push
