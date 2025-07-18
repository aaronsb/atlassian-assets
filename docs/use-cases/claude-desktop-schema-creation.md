# Claude Desktop Schema Creation Use Case

## Overview

This document demonstrates how Claude Desktop can use the Atlassian Assets MCP server to create complex schema hierarchies through natural conversation, transforming workshop discussions into structured asset management systems.

## Real-World Scenario: Digital Content Creation System

### The Challenge

A creative agency needs to organize their digital assets across multiple projects. They want to create a system that can handle:

- **Visual Content**: Images, videos, 3D models
- **Audio Content**: Music tracks, sound effects, podcast episodes
- **Hierarchical Organization**: Abstract categories with concrete instances
- **Searchable Structure**: Easy discovery across all content types

### The Solution: Conversational Schema Creation

Instead of manually clicking through Atlassian Assets UI, the team described their needs to Claude Desktop, which used the MCP server to create the entire system.

## Step-by-Step Creation Process

### 1. Initial Conversation
**User**: "I need to create a digital content management system for our creative projects"

**Claude Desktop**: Uses `assets_browse_schema` to understand the current structure, then plans the hierarchy.

### 2. Schema Architecture Design
Claude Desktop created a sophisticated hierarchy:

```
Digital Asset (Root Abstract Type)
├── Visual Content (Abstract)
│   ├── Image (Concrete)
│   ├── Video (Concrete)
│   └── 3D Model (Concrete)
└── Audio Content (Abstract)
    ├── Music Track (Concrete)
    ├── Sound Effect (Concrete)
    └── Podcast Episode (Concrete)
```

### 3. Object Type Creation
Using `assets_create_object_type`, Claude Desktop systematically created:

- **3 Abstract Types**: Digital Asset, Visual Content, Audio Content
- **6 Concrete Types**: Image, Video, 3D Model, Music Track, Sound Effect, Podcast Episode
- **Total**: 9 object types with proper parent-child relationships

### 4. Sample Content Population
Claude Desktop then populated the schema with realistic examples:

| Type | Object Name | Key | Purpose |
|------|-------------|-----|---------|
| Image | Hero Banner - Summer Campaign 2025 | CLITEST202-1025 | Marketing visual |
| Video | Product Demo Video - AI Assistant | CLITEST202-1026 | Product showcase |
| 3D Model | Spaceship Model - Low Poly Game Asset | CLITEST202-1027 | Game development |
| Music Track | Epic Orchestra Theme - Battle Scene | CLITEST202-1028 | Game soundtrack |
| Sound Effect | UI Click Sound - Soft Pop | CLITEST202-1029 | Interface audio |
| Podcast Episode | Tech Talk Episode 42 - The Future of AI | CLITEST202-1030 | Content series |

### 5. Verification and Visualization
Claude Desktop used `assets_search` to verify all created objects and provided a comprehensive summary with Mermaid diagrams.

## Key Benefits Demonstrated

### 1. **Conversational Interface**
- No need to learn Atlassian Assets UI
- Natural language describes complex requirements
- Instant schema creation from business needs

### 2. **Intelligent Structure**
- Proper abstract/concrete type relationships
- Automatic attribute generation (Key, Name, Created, Updated)
- Hierarchical organization for easy navigation

### 3. **Complete Workflow**
- Schema design → Object type creation → Sample data → Verification
- All done through natural conversation
- Real-time feedback and adjustments

### 4. **Production-Ready Results**
- Schema immediately usable for asset management
- Searchable across all content types
- Extensible for future requirements

## Workshop Integration Potential

### Discovery Session → Schema Creation
Imagine a workshop where stakeholders describe their asset management needs:

**Stakeholder**: *"We have marketing images, product videos, and we're starting a podcast. We also do game development with 3D models and sound effects. Everything needs to be searchable and organized by project."*

**Claude Desktop**: *"I'll create a Digital Content Creation system for you with proper hierarchies..."*

Within minutes, the conversation results in a complete, functional schema in Atlassian Assets.

### Call Transcript Integration
Future enhancements could include:
- **Audio Transcript Processing**: Upload workshop recordings
- **Automatic Schema Generation**: Extract requirements from conversation
- **Iterative Refinement**: Adjust schema based on feedback
- **Multi-Stakeholder Input**: Incorporate different perspectives

## Technical Implementation

### MCP Tools Used
- `assets_browse_schema` - Understanding existing structure
- `assets_create_object_type` - Creating type hierarchy
- `assets_create_object` - Populating with sample data
- `assets_search` - Verification and discovery
- `assets_get` - Detailed object inspection

### Workflow Benefits
- **Speed**: Complete schema in minutes vs. hours
- **Accuracy**: Proper relationships and attributes
- **Consistency**: Standardized naming and structure
- **Documentation**: Auto-generated summaries and diagrams

## Screenshots Reference

The following screenshots document the complete process:

1. **`claude-desktop-creating-schema-entities.jpeg`**: Shows Claude Desktop creating diverse object types through conversation
2. **`new-schema-objects.jpeg`**: Displays the created schema tree in Atlassian Assets UI
3. **`schema-creation-summary.jpeg`**: Claude Desktop's comprehensive summary of what was built
4. **`schema-review-diagram.jpeg`**: Mermaid diagram showing the complete hierarchy and relationships

## Conclusion

This use case demonstrates the power of combining conversational AI with structured asset management. The MCP server enables Claude Desktop to translate natural language requirements into production-ready schemas, dramatically reducing the time and expertise needed for complex system setup.

The workflow transforms asset management from a technical configuration task into a natural conversation about business needs.