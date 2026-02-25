# ai content
CURSOR_RULES_INIT_URL := https://raw.githubusercontent.com/hicker-kin/ai-context/main/cursor_rules.sh
CURSOR_INIT := scripts/cursor_rules.sh

CLAUDE_RULES_INIT_URL := https://raw.githubusercontent.com/hicker-kin/ai-context/main/claude_rules.sh
CLAUDE_INIT := scripts/claude_rules.sh

# install cursor init scripts
install-cursor-rule-init:
	@echo "Downloading cursor rules init scripts..."
	@mkdir -p scripts
	@curl -fL $(CURSOR_RULES_INIT_URL) -o $(CURSOR_INIT)
	@chmod +x $(CURSOR_INIT)
	@echo "Done -> $(CURSOR_INIT)"

install-claude-rule-init:
	@echo "Downloading claude rules init scripts..."
	@mkdir -p scripts
	@curl -fL $(CLAUDE_RULES_INIT_URL) -o $(CLAUDE_INIT)
	@chmod +x $(CLAUDE_INIT)
	@echo "Done -> $(CLAUDE_INIT)"

ai-rules-init: install-cursor-rule-init install-claude-rule-init

cursor-go-rules:
	@echo "Generating cursor rules..."
	@sh scripts/cursor_rules.sh go

claude-go-rules:
	@echo "Generating claude rules..."
	@sh scripts/claude_rules.sh go

# install ai rules
ai-rules-install: ai-rules-init cursor-go-rules claude-go-rules
