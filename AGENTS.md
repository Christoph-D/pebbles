# AGENTS.md - Code Style & Development Guide

## Go Code Style

### Structs & Interfaces
- Use dependency injection via concrete types unless you have multiple implementations
- Don't create interfaces unless you have multiple implementations or need to mock for tests
- Keep structs simple, focus on single responsibility
- Use pointer receivers for methods that modify state
- Use value receivers for methods that don't modify state
