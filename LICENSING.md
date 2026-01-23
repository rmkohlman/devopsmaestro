# DevOpsMaestro Licensing Guide

DevOpsMaestro uses a **dual-license model** to remain free for individual developers while generating revenue from corporate users.

## 📋 Quick Decision Guide

### ✅ FREE (GPL-3.0) - No License Needed

You can use DevOpsMaestro **completely free** if you are:

- 👤 **Individual developer** working on personal projects
- 🎓 **Student or educator** using it for learning/teaching
- 🌍 **Open source developer** creating GPL-compatible software
- 🏛️ **Non-profit organization**
- 🧪 **Researcher** using it for academic purposes
- 🏠 **Hobbyist** building side projects

**No registration, no payment, no restrictions!**

### 💼 PAID (Commercial License) - License Required

You **must purchase** a commercial license if:

- 🏢 You are a **corporation, LLC, business entity**
- 💰 You use DevOpsMaestro for **commercial development**
- 👥 Your **employees use** DevOpsMaestro at work
- 🔒 You want to build **proprietary/closed-source** software
- 🚫 You cannot comply with **GPL-3.0 copyleft** requirements
- ⚡ You want **priority support** and custom features

---

## 🆓 GPL-3.0 License (Free Tier)

### What You Get

- ✅ Full source code access
- ✅ Unlimited personal use
- ✅ Modify and extend freely
- ✅ Use for open source projects
- ✅ Community support
- ✅ All features and updates

### Requirements

- Must keep code **open source** (GPL-3.0 compatible)
- Must share modifications under **same license**
- Must provide **source code** to users
- Cannot use in **proprietary software** (without commercial license)

### Perfect For

```bash
# Individual developers
$ dvm init  # Use for personal projects - FREE!

# Open source projects
$ git clone your-oss-project
$ dvm workspace create main
$ dvm build  # All features, no restrictions!

# Learning and education
$ dvm plugin list  # Study how it works, modify it!
```

See [LICENSE](LICENSE) for full GPL-3.0 terms.

---

## 💼 Commercial License (Paid Tier)

### What You Get

- ✅ **No GPL restrictions** - use in proprietary software
- ✅ **Private modifications** - keep changes private
- ✅ **Priority support** - email/chat support with SLA
- ✅ **Legal protection** - indemnification and warranties
- ✅ **Custom features** - request enterprise features
- ✅ **Team licenses** - cover all your developers

### Pricing Tiers

#### 🏢 Small Team License
- **Up to 10 developers**
- Annual subscription: Contact for pricing
- Email support (48h response)
- All features included

#### 🏭 Enterprise License
- **Unlimited developers**
- Annual or perpetual options
- Priority support with SLA
- Custom integrations available
- Volume discounts available

#### 🔧 OEM/Reseller License
- **Embed in your product**
- White-label options
- Royalty or flat-fee pricing
- Dedicated account manager

### How to Purchase

**Contact us:**
- 📧 Email: support@devopsmaestro.io
- 🐙 GitHub: https://github.com/rmkohlman/devopsmaestro/issues
- 💬 Start a discussion: [GitHub Discussions](https://github.com/rmkohlman/devopsmaestro/discussions)

**We'll need:**
- Company name and size
- Number of developers
- Use case description
- License type preference (annual vs perpetual)

See [LICENSE-COMMERCIAL.txt](LICENSE-COMMERCIAL.txt) for detailed terms.

---

## ❓ Frequently Asked Questions

### Q: I'm a freelancer. Do I need a commercial license?

**A:** If you're operating as an **individual** (sole proprietor), you can use the free GPL-3.0 license. If you've incorporated your business (LLC, Inc), you need a commercial license.

### Q: Can I try it before buying?

**A:** Yes! Use the GPL-3.0 version first. When you're ready to use it commercially, purchase a license.

### Q: What if I'm a startup with no revenue?

**A:** Contact us! We offer **discounts for startups** and early-stage companies. We want to support innovation.

### Q: Can I use GPL-3.0 at a company for internal tools?

**A:** Technically yes, **IF** you're willing to open-source your entire internal tooling under GPL-3.0. Most companies prefer the commercial license to avoid this.

### Q: What happens if my team grows beyond 10 developers?

**A:** Just upgrade to Enterprise License. We'll prorate your existing subscription.

### Q: Do I need a license for CI/CD servers?

**A:** No! Build servers, CI/CD pipelines, and deployment automation don't count toward developer seats.

### Q: Can I fork and modify under GPL-3.0?

**A:** Yes! Fork freely, modify, and share. Just keep it GPL-3.0 licensed and open source.

### Q: What if I want to contribute back to DevOpsMaestro?

**A:** Contributions are welcome! By contributing, you agree to dual-license your contributions (GPL-3.0 + our commercial terms).

---

## 📊 Comparison Table

| Feature | GPL-3.0 (Free) | Commercial License |
|---------|----------------|-------------------|
| **Personal Use** | ✅ Unlimited | ✅ Unlimited |
| **Commercial Use** | ⚠️ Must open-source | ✅ Proprietary OK |
| **Modify Source** | ✅ Yes (must share) | ✅ Yes (keep private) |
| **All Features** | ✅ Yes | ✅ Yes |
| **Updates** | ✅ Yes | ✅ Yes |
| **Community Support** | ✅ Yes | ✅ Yes |
| **Priority Support** | ❌ No | ✅ Yes (SLA) |
| **Legal Protection** | ❌ No | ✅ Yes |
| **Custom Features** | ❌ No | ✅ Available |
| **Team Licenses** | ❌ No | ✅ Yes |
| **White-label** | ❌ No | ✅ Available (OEM) |

---

## 🎯 Examples

### ✅ Free GPL-3.0 Use Cases

```bash
# Personal side project
$ cd my-personal-app
$ dvm init
$ dvm workspace create main
# 100% free, use all features!

# Open source project
$ cd my-oss-framework  # Project is GPL-3.0 licensed
$ dvm plugin apply -f telescope.yaml
# Perfect! Both are GPL-3.0

# Learning/education
$ dvm version
$ dvm get plugins
# Study, learn, experiment - completely free!
```

### 💼 Requires Commercial License

```bash
# At your company
Company LLC$ dvm init
# ⚠️ Need commercial license - contact us!

# Proprietary SaaS product
Startup Inc$ dvm workspace create prod
# ⚠️ Need commercial license for closed-source

# Client consulting work
Consulting Co$ dvm build
# ⚠️ If incorporated, need commercial license
```

---

## 🤝 Why Dual-License?

**For Individual Developers:**
- Keep it FREE for personal use
- Foster learning and innovation
- Build open source community
- No barriers to entry

**For Companies:**
- Support ongoing development
- Get priority support and features
- Avoid GPL compliance burden
- Legal protection and warranties

**For Everyone:**
- Sustainable business model
- Faster feature development
- Better support and documentation
- Active maintenance and updates

---

## 📞 Contact & Support

**General Questions:**
- GitHub Issues: https://github.com/rmkohlman/devopsmaestro/issues
- GitHub Discussions: https://github.com/rmkohlman/devopsmaestro/discussions

**Commercial Licensing:**
- Email: support@devopsmaestro.io
- Include: Company name, team size, use case

**Contributing:**
- See [CONTRIBUTING.md](CONTRIBUTING.md)
- All contributions dual-licensed (GPL-3.0 + Commercial)

---

## 📜 License Files

- [LICENSE](LICENSE) - Full GPL-3.0 text (free tier)
- [LICENSE-COMMERCIAL.txt](LICENSE-COMMERCIAL.txt) - Commercial license terms
- This file (LICENSING.md) - Human-friendly guide

---

**Made with ❤️ by developers, for developers.**

Free for individuals, sustainable through corporate support.
