# AI Generation Guide

Leverage OpenAI's powerful language models to generate professional content directly from the command line.

## 🤖 Overview

The `dox generate` command uses OpenAI's GPT models to create various types of content based on your prompts. This guide covers setup, usage, and best practices for AI-powered content generation.

## 🔑 Setup

### Prerequisites

1. **OpenAI Account**: Create an account at [OpenAI](https://platform.openai.com)
2. **API Key**: Generate an API key from the [API Keys page](https://platform.openai.com/api-keys)
3. **Credits**: Ensure you have sufficient API credits

### Configuration

#### Method 1: Environment Variable
```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
export OPENAI_API_KEY="sk-your-api-key-here"
```

#### Method 2: Config File
```bash
# Set via dox config
dox config --set openai.api_key="sk-your-api-key-here"
```

#### Method 3: Command Flag
```bash
# Pass directly in command
dox generate --api-key "sk-your-api-key-here" ...
```

### Verify Setup
```bash
# Test with a simple generation
dox generate --type summary --prompt "Test" --output test.md
```

## 📝 Content Types

### Blog Posts
Generate engaging blog content:
```bash
dox generate --type blog \
  --prompt "5 ways AI is transforming healthcare" \
  --output healthcare-ai-blog.md
```

**Optimized for**:
- SEO-friendly structure
- Engaging introduction
- Clear sections with headers
- Conclusion with call-to-action

### Business Reports
Create professional reports:
```bash
dox generate --type report \
  --prompt "Q4 2024 sales performance analysis with recommendations" \
  --output q4-sales-report.md \
  --temperature 0.3
```

**Features**:
- Executive summary
- Data-driven insights
- Clear recommendations
- Professional tone

### Summaries
Condensed content from longer texts:
```bash
# Summarize a document
cat long-document.md | dox generate --type summary \
  --prompt "Summarize this document focusing on key findings" \
  --output summary.md
```

**Characteristics**:
- Concise bullet points
- Key takeaways highlighted
- Maintains accuracy
- Preserves important details

### Professional Emails
Draft business communications:
```bash
dox generate --type email \
  --prompt "Request meeting with client about project delays" \
  --output meeting-request.md \
  --temperature 0.5
```

**Includes**:
- Appropriate greeting
- Clear subject/purpose
- Professional tone
- Action items

### Business Proposals
Create compelling proposals:
```bash
dox generate --type proposal \
  --prompt "Web development services for e-commerce platform" \
  --output proposal.md \
  --max-tokens 3000
```

**Structure**:
- Executive summary
- Scope of work
- Timeline
- Budget considerations
- Next steps

### Custom Content
Full control over generation:
```bash
dox generate --type custom \
  --prompt "Write a technical guide for implementing OAuth 2.0 in Node.js" \
  --output oauth-guide.md \
  --model gpt-4 \
  --max-tokens 4000
```

## ⚙️ Parameters

### Model Selection
Choose the AI model:
```bash
# Default: GPT-3.5 Turbo (fast, cost-effective)
dox generate --model gpt-3.5-turbo ...

# GPT-4 (more capable, higher cost)
dox generate --model gpt-4 ...

# GPT-4 Turbo (best quality)
dox generate --model gpt-4-turbo-preview ...
```

### Temperature
Control creativity (0.0 to 2.0):
```bash
# Low (0.0-0.3): Focused, deterministic
dox generate --temperature 0.2 ...  # Best for reports, documentation

# Medium (0.4-0.7): Balanced
dox generate --temperature 0.5 ...  # Good for emails, summaries

# High (0.8-2.0): Creative, varied
dox generate --temperature 1.0 ...  # Best for blogs, creative content
```

### Max Tokens
Control response length:
```bash
# Short content (500-1000 tokens)
dox generate --max-tokens 500 ...   # ~375 words

# Medium content (1000-2000 tokens)
dox generate --max-tokens 1500 ...  # ~1125 words

# Long content (2000-4000 tokens)
dox generate --max-tokens 3000 ...  # ~2250 words
```

### Language
Generate in different languages:
```bash
# Korean
dox generate --language Korean ...

# Spanish
dox generate --language Spanish ...

# Japanese
dox generate --language Japanese ...
```

## 🎯 Advanced Usage

### Multi-Step Generation
Build complex documents:
```bash
# 1. Generate outline
dox generate --type custom \
  --prompt "Create an outline for a guide about Docker" \
  --output outline.md

# 2. Expand each section
dox generate --type custom \
  --prompt "Expand on Docker installation section with examples" \
  --output installation.md

# 3. Combine and format
cat outline.md installation.md | dox create --from - --output guide.docx
```

### Template Integration
Combine AI generation with templates:
```bash
# 1. Generate content
dox generate --type report \
  --prompt "Monthly marketing metrics analysis" \
  --output content.md

# 2. Convert to data format
echo "content: |" > data.yml
cat content.md | sed 's/^/  /' >> data.yml

# 3. Fill template
dox template -t report-template.docx -v data.yml -o final-report.docx
```

### Batch Generation
Generate multiple pieces:
```bash
# Create a prompts file
cat > prompts.txt << EOF
Blog: Cloud computing trends 2025
Blog: Remote work best practices
Blog: Cybersecurity for small business
EOF

# Generate all
while IFS=': ' read type prompt; do
  slug=$(echo "$prompt" | tr ' ' '-' | tr '[:upper:]' '[:lower:]')
  dox generate --type blog --prompt "$prompt" --output "blogs/${slug}.md"
done < prompts.txt
```

### Context Enhancement
Provide additional context:
```bash
# Include context file
dox generate --type custom \
  --prompt "Based on the following data, write an analysis: $(cat data.txt)" \
  --output analysis.md

# Use system message for context
dox generate --type report \
  --system "You are a financial analyst specializing in tech stocks" \
  --prompt "Analyze AAPL Q4 performance" \
  --output apple-analysis.md
```

## 💰 Cost Optimization

### Token Usage
Understand pricing:
- **GPT-3.5 Turbo**: ~$0.002 per 1K tokens
- **GPT-4**: ~$0.03 per 1K tokens
- **GPT-4 Turbo**: ~$0.01 per 1K tokens

### Optimization Tips

1. **Use appropriate models**
```bash
# Use GPT-3.5 for simple tasks
dox generate --model gpt-3.5-turbo --type summary ...

# Use GPT-4 only when needed
dox generate --model gpt-4 --type custom --prompt "Complex technical analysis" ...
```

2. **Limit token usage**
```bash
# Set reasonable limits
dox generate --max-tokens 1000 ...  # Usually sufficient
```

3. **Cache responses**
```bash
# Enable caching
dox config --set performance.cache=true
```

4. **Batch similar requests**
```bash
# Combine related prompts
dox generate --type custom \
  --prompt "Write 3 blog post titles about: 1) AI 2) Cloud 3) Security" \
  --output titles.md
```

## 📊 Output Formats

### Markdown (Default)
```bash
dox generate --format markdown ...
```

### Plain Text
```bash
dox generate --format text ...
```

### JSON
```bash
dox generate --format json \
  --prompt "Generate product descriptions" \
  --output products.json
```

### HTML
```bash
dox generate --format html \
  --prompt "Create a landing page" \
  --output landing.html
```

### Direct to Word/PowerPoint
```bash
# Generate and convert
dox generate --type blog --prompt "..." --output temp.md && \
dox create --from temp.md --output blog.docx && \
rm temp.md
```

## 🎨 Prompt Engineering

### Effective Prompts

#### Be Specific
```bash
# Too vague
dox generate --prompt "Write about technology"

# Better
dox generate --prompt "Write a 500-word blog post about how 5G technology will impact remote work in 2025, focusing on speed improvements and new possibilities"
```

#### Provide Structure
```bash
dox generate --type custom \
  --prompt "Write a report with: 1) Executive Summary 2) Current State Analysis 3) Recommendations 4) Implementation Timeline"
```

#### Define Audience
```bash
dox generate --prompt "Explain blockchain technology for non-technical business executives"
```

#### Specify Tone
```bash
dox generate --prompt "Write a friendly, conversational blog post about productivity tips for remote workers"
```

### Prompt Templates

#### Blog Post
```bash
PROMPT="Write a [word count]-word blog post about [topic].
Target audience: [audience]
Tone: [tone]
Include: [specific points]
Call-to-action: [CTA]"
```

#### Report
```bash
PROMPT="Create a business report analyzing [subject].
Data points: [data]
Time period: [period]
Key metrics: [metrics]
Recommendations: Yes/No"
```

#### Email
```bash
PROMPT="Draft a professional email to [recipient] regarding [subject].
Purpose: [purpose]
Key points: [points]
Desired outcome: [outcome]
Tone: [formal/friendly/urgent]"
```

## 🚨 Error Handling

### Common Issues

#### API Key Invalid
```bash
Error: Invalid API key
Solution: Verify your API key is correct and active
dox config --set openai.api_key="sk-correct-key"
```

#### Rate Limiting
```bash
Error: Rate limit exceeded
Solution: Wait and retry, or upgrade your OpenAI plan
# Add delay between requests
sleep 2
```

#### Token Limit Exceeded
```bash
Error: Maximum token limit exceeded
Solution: Reduce max_tokens or split into smaller requests
dox generate --max-tokens 2000 ...  # Instead of 4000
```

#### Network Issues
```bash
Error: Connection timeout
Solution: Check internet connection, retry with longer timeout
dox generate --timeout 60 ...
```

## 🔒 Security Best Practices

### API Key Management
1. **Never commit API keys** to version control
2. **Use environment variables** in production
3. **Rotate keys regularly**
4. **Set usage limits** in OpenAI dashboard

### Sensitive Data
1. **Avoid sending** confidential information in prompts
2. **Review generated content** before sharing
3. **Use local models** for sensitive data (future feature)

### Compliance
1. **Check regulations** for AI-generated content in your industry
2. **Add disclaimers** when appropriate
3. **Maintain human oversight** for critical content

## 📈 Monitoring Usage

### Track API Usage
```bash
# View current month usage (via OpenAI dashboard)
# https://platform.openai.com/usage
```

### Log Generation History
```bash
# Enable logging
dox config --set logging.enabled=true
dox config --set logging.path="./logs"

# View logs
cat logs/generation.log
```

### Cost Tracking
```bash
# Estimate costs before generation
dox generate --estimate-cost --max-tokens 2000 --model gpt-4
# Estimated cost: $0.06
```

## 🎯 Best Practices

### Quality Control
1. **Review all output** before using
2. **Fact-check** important information
3. **Edit for brand voice** and style
4. **Test with small samples** first

### Workflow Integration
1. **Create templates** for common prompts
2. **Automate repetitive** generation tasks
3. **Combine with other** dox features
4. **Version control** important outputs

### Performance
1. **Cache frequently used** responses
2. **Batch similar requests**
3. **Use appropriate models** for each task
4. **Monitor and optimize** token usage

## 📚 Examples Library

### Marketing Content
```bash
# Product description
dox generate --type custom \
  --prompt "Write a compelling product description for eco-friendly water bottle" \
  --temperature 0.8 \
  --output product-desc.md

# Social media posts
dox generate --type custom \
  --prompt "Create 5 Twitter posts promoting our new AI course" \
  --max-tokens 500 \
  --output social-posts.md
```

### Technical Documentation
```bash
# API documentation
dox generate --type custom \
  --prompt "Document REST API endpoints for user management system" \
  --temperature 0.2 \
  --output api-docs.md

# README file
dox generate --type custom \
  --prompt "Create a README.md for Python web scraping library" \
  --output README.md
```

### Business Documents
```bash
# Meeting agenda
dox generate --type custom \
  --prompt "Create agenda for quarterly planning meeting" \
  --output agenda.md

# Executive brief
dox generate --type report \
  --prompt "Executive brief on market expansion opportunities in Asia" \
  --max-tokens 2000 \
  --output brief.md
```

## 🔗 Next Steps

- Explore [Templates Guide](./templates.md) to combine AI with templates
- Check [Command Reference](./commands.md) for all options
- See [Examples](../examples/) for real-world scenarios
- Learn about [Configuration](./configuration.md) options