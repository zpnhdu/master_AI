package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Skill struct {
	Name          string
	Path          string
	Description   string
	RawContent    string
	HasReferences bool
	HasScripts    bool
}

type Loader struct {
	skillDir string
}

func NewLoader(skillDir string) *Loader {
	return &Loader{skillDir: skillDir}
}

func (l *Loader) ListSkills() []Skill {
	skillsByName := make(map[string]Skill)
	for _, skill := range defaultSkills() {
		skillsByName[skill.Name] = skill
	}

	for _, skill := range l.scanDiskSkills() {
		skillsByName[skill.Name] = skill
	}

	out := make([]Skill, 0, len(skillsByName))
	for _, skill := range skillsByName {
		out = append(out, skill)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func (l *Loader) LoadSkill(name string) (*Skill, error) {
	for _, skill := range l.ListSkills() {
		if skill.Name == name {
			return &skill, nil
		}
	}
	return nil, fmt.Errorf("skill %s not found", name)
}

func BuildSkillPrompt(skills []Skill) string {
	if len(skills) == 0 {
		return "未加载额外 Skill，请按通用科研与求职助理方法回答。"
	}

	var b strings.Builder
	b.WriteString("以下 Skill 只提供任务方法论，不直接执行外部操作：\n")
	for _, skill := range skills {
		b.WriteString("\n## ")
		b.WriteString(skill.Name)
		if skill.Description != "" {
			b.WriteString("\n")
			b.WriteString(skill.Description)
		}
		b.WriteString("\n")
		b.WriteString(skill.RawContent)
		b.WriteString("\n")
	}
	return b.String()
}

func (l *Loader) scanDiskSkills() []Skill {
	if l == nil || l.skillDir == "" {
		return nil
	}
	info, err := os.Stat(l.skillDir)
	if err != nil || !info.IsDir() {
		return nil
	}

	var skills []Skill
	_ = filepath.WalkDir(l.skillDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.ToLower(d.Name()) != "skill.md" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		dir := filepath.Dir(path)
		name := filepath.Base(dir)
		if name == "." || name == string(filepath.Separator) {
			name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
		raw := string(content)
		skills = append(skills, Skill{
			Name:          name,
			Path:          path,
			Description:   extractDescription(raw),
			RawContent:    raw,
			HasReferences: dirExists(filepath.Join(dir, "references")),
			HasScripts:    dirExists(filepath.Join(dir, "scripts")),
		})
		return nil
	})
	return skills
}

func extractDescription(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func defaultSkills() []Skill {
	return []Skill{
		defaultSkill("paper-reading", "按 Background / Problem / Method / Experiment / Result / Limitation 总结论文。", `阅读论文时先识别研究背景、核心问题、方法设计、实验设置、结果结论和局限性。回答时优先说明论文解决了什么问题、为什么重要、方法的关键假设、实验是否支撑结论，以及后续可扩展方向。`),
		defaultSkill("literature-review", "提取关键词后搜索相关论文，并按方法、数据、结果、局限进行对比。", `做文献综述时先抽取 3-8 个核心关键词，再按方法路线、数据集、评价指标、结果质量和局限性组织材料。避免简单罗列论文，优先给出脉络、差异和可复用结论。`),
		defaultSkill("code-reproduction", "根据论文方法搜索 GitHub 仓库，判断 README、环境、数据、脚本和复现难度。", `评估代码复现时检查仓库活跃度、依赖环境、数据获取方式、训练/推理脚本、模型权重、issue 质量和许可证。输出应包含复现路径、主要风险和最小可运行验证方案。`),
		defaultSkill("research-presentation", "生成组会汇报大纲，覆盖背景、问题、方法、结果、不足和下一步计划。", `组会汇报应从研究背景切入，快速定义问题和动机，再讲方法核心、实验结果、失败/不足、下一步计划。注意把技术细节翻译成听众能理解的研究价值。`),
		defaultSkill("resume-optimization", "根据 JD 和用户经历优化简历项目描述，突出技术栈、工程难点、量化结果和岗位相关性。", `优化简历时先拆 JD 能力点，再把用户经历映射到技术栈、工程难点、业务效果和量化指标。输出要适合直接改写到简历，避免夸大未实现能力。`),
		defaultSkill("interview-project", "按项目背景、系统架构、核心模块、技术难点、优化点、可追问问题组织项目讲解。", `准备项目面试时按背景目标、架构设计、核心链路、技术难点、取舍、优化和追问组织。回答应能体现工程判断，而不是只背功能清单。`),
	}
}

func defaultSkill(name, description, content string) Skill {
	return Skill{
		Name:        name,
		Path:        "builtin://" + name,
		Description: description,
		RawContent:  content,
	}
}
