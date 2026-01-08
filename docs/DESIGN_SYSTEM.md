# Minimax Voice Workbench - Design System Specification

## 1. 概述
本设计系统旨在为 Minimax Voice Workbench 提供统一、现代化且易于维护的视觉语言。系统基于 CSS Variables 构建，原生支持明暗双色主题。

## 2. 色彩系统 (Color Palette)

### 品牌色 (Brand Colors)
| Token | Light Mode | Dark Mode | 用途 |
|-------|------------|-----------|------|
| `--primary` | `#6366f1` (Indigo 500) | `#818cf8` (Indigo 400) | 主要操作、激活状态 |
| `--primary-hover` | `#4f46e5` | `#6366f1` | 悬停状态 |
| `--primary-bg` | `#e0e7ff` | `rgba(99, 102, 241, 0.15)` | 浅色背景、选中项 |

### 功能色 (Functional Colors)
| Token | Hex (Light) | 用途 |
|-------|-------------|------|
| `--success` | `#10b981` | 成功状态、完成 |
| `--warning` | `#f59e0b` | 警告、处理中 |
| `--error` | `#ef4444` | 错误、危险操作 |
| `--info` | `#3b82f6` | 信息提示、系统音色 |

### 中性色 (Neutrals)
| Token | Light Mode | Dark Mode | 用途 |
|-------|------------|-----------|------|
| `--bg-primary` | `#ffffff` | `#0f172a` | 页面背景 |
| `--bg-secondary` | `#f8fafc` | `#1e293b` | 卡片背景、侧边栏 |
| `--bg-tertiary` | `#f1f5f9` | `#334155` | 输入框背景、分割线 |
| `--text-primary` | `#0f172a` | `#f8fafc` | 主要标题、正文 |
| `--text-secondary` | `#475569` | `#cbd5e1` | 次要信息、标签 |
| `--text-tertiary` | `#94a3b8` | `#94a3b8` | 提示文本、占位符 |

## 3. 排版 (Typography)

字体家族：`'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`

| 等级 | 大小 (rem/px) | 字重 | 行高 | 用途 |
|------|---------------|------|------|------|
| H1 | 2.25 / 36 | 700 | 1.2 | 页面主标题 |
| H2 | 1.875 / 30 | 700 | 1.2 | 模块/卡片标题 |
| H3 | 1.5 / 24 | 600 | 1.3 | 子模块标题 |
| Body | 1 / 16 | 400 | 1.6 | 正文内容 |
| Small | 0.875 / 14 | 400 | 1.5 | 辅助说明、标签 |

## 4. 间距与布局 (Spacing & Layout)

基于 4px 网格系统：
*   `--space-1`: 4px
*   `--space-2`: 8px
*   `--space-3`: 12px
*   `--space-4`: 16px (标准间距)
*   `--space-6`: 24px (卡片内边距)
*   `--space-8`: 32px
*   `--space-10`: 40px

## 5. 组件规范 (Components)

### 卡片 (Card)
*   背景：`--bg-secondary`
*   圆角：`--radius-lg` (12px)
*   边框：`1px solid --border-color`
*   阴影：`--shadow-sm` (Hover 时 `--shadow-md`)

### 按钮 (Button)
*   高度：40px (Default), 32px (Small), 48px (Large)
*   圆角：`--radius-md` (8px)
*   Padding: horizontal 20px
*   变体：Primary (实色), Secondary (描边/浅背景), Ghost (无背景)

### 表单 (Form)
*   输入框高度：40px
*   圆角：`--radius-md`
*   背景：`--bg-primary` (Light) / `--bg-primary` (Dark)
*   Focus 状态：`--primary` 边框 + 3px `primary-bg` 阴影环

## 6. 交互规范 (Interaction)

*   **过渡动画**：所有可交互元素具备 `transition: all 150ms ease`。
*   **反馈**：
    *   按钮点击有 Active 缩放效果 (`transform: translateY(0)` vs Hover `translateY(-1px)`).
    *   耗时操作显示 Loading Spinner。
*   **响应式**：
    *   Desktop (>1024px): 完整布局，双栏/多栏。
    *   Tablet (768-1024px): 侧边栏折叠或自适应 Grid。
    *   Mobile (<768px): 侧边栏隐藏（汉堡菜单），单栏布局，触控区域优化 (min 44px)。
