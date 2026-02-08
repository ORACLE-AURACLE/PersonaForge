# TODO: Integrate Custom Markdown Parser into Chat and Insights

## Tasks
- [x] Filter markdown.jsx: Remove table parsing, code block handling, and blockquotes to simplify for chat/insights workflow.
- [x] Update persona-forge/src/pages/PersonaChat.jsx: Replace ReactMarkdown import and usage with parseMarkdown from custom markdown.jsx.
- [x] Update persona-forge/src/components/PersonaChat.jsx: Replace ReactMarkdown import and usage with parseMarkdown from custom markdown.jsx.
- [x] Test rendering in chat and insights panels to ensure markdown displays correctly.
- [x] Verify no console errors and that security utils work properly.
