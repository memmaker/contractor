package d100

import (
	"bytes"
	"cmp"
	"encoding/gob"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"slices"
	"strings"
)

// Requirements
// Factor in modifiers to all these stats by..
// 1. Equipment
// 2. Perks
// 3. Traits
// 4. Drugs
// 5. Environment
// 6. Status effects
// 7. Party members

func (cs *CharSheet) getSkillBase(skill Skill) int {
	return skill.BaseValue(nil)
}

// PrintReport: If the task at hand is simply not possible for someone without a certain level of skill
// Roll: If the task at hand is in general possible, even if it's difficult

// tagging gives 20% bonus to skill
func NewCharSheet() *CharSheet {
	c := &CharSheet{
		derivedStatAdjustments: make(map[DerivedStat]int),
		skillAdjustments:       make(map[Skill]int),
		taggedSkills:           make(map[Skill]bool),
	}
	c.HealAPAndHPCompletely()
	return c
}

type PerkLevel struct {
	Perk  Perk
	Level int
}
type CharSheet struct {
	availableSkillPoints int
	availablePerks       int

	perks []PerkLevel

	derivedStatAdjustments map[DerivedStat]int
	skillAdjustments       map[Skill]int

	taggedSkills map[Skill]bool

	hitPointsCurrent int

	actionPointsCurrent int

	getDerivedStatMods func(DerivedStat) []Modifier
	getSkillMods       func(Skill) []Modifier

	onDerivedStatChangedHandler func(DerivedStat)
	onSkillChangedHandler       func(Skill)
}

type FactorModifier struct {
	Source         string
	ModifierFactor float64
	Order          int
	Suffix         string
}

func (f FactorModifier) Description() string {
	var line string
	line = fmt.Sprintf("%s: x%.2f", f.Source, f.ModifierFactor)
	if f.Suffix != "" {
		line += " " + f.Suffix
	}
	return line
}

func (f FactorModifier) Apply(i int) int {
	return int(float64(i) * f.ModifierFactor)
}

func (f FactorModifier) SortOrder() int {
	return f.Order
}

func (f FactorModifier) IsNonModifying() bool {
	return f.ModifierFactor == 1.0
}

func (f FactorModifier) ApplyForInterval(value fxtools.Interval) fxtools.Interval {
	return fxtools.Interval{Min: f.Apply(value.Min), Max: f.Apply(value.Max)}
}

type DefaultModifier struct {
	Source    string
	Modifier  int
	Order     int
	IsPercent bool
	Suffix    string
}

func (d DefaultModifier) ApplyForInterval(value fxtools.Interval) fxtools.Interval {
	return fxtools.Interval{Min: d.Apply(value.Min), Max: d.Apply(value.Max)}
}

func (d DefaultModifier) Description() string {
	var line string
	if d.IsPercent {
		line = fmt.Sprintf("%s: %+d%%", d.Source, d.Modifier)
	}
	line = fmt.Sprintf("%s: %+d", d.Source, d.Modifier)
	if d.Suffix != "" {
		line += " " + d.Suffix
	}
	return line
}

func (d DefaultModifier) Apply(i int) int {
	return i + d.Modifier
}

func (d DefaultModifier) SortOrder() int {
	return d.Order
}
func (d DefaultModifier) IsNonModifying() bool {
	return d.Modifier == 0
}

var QuickDrawModifier = FactorModifier{
	Source:         "Quick Draw",
	ModifierFactor: 1.5,
	Order:          0,
}

var NoModifierList []Modifier

var NoCombatModifier = CombatModifiers{}

type CombatModifiers struct {
	ChanceToHitMods ModList
	DamageMods      ModList
}

func (m CombatModifiers) WithChanceToHit(modifiers ModList) CombatModifiers {
	m.ChanceToHitMods = append(m.ChanceToHitMods, modifiers...)
	return m
}

type ModList []Modifier

func (r ModList) appendIfNonZero(mod Modifier) ModList {
	if mod.IsNonModifying() {
		return r
	}
	return append(r, mod)
}

func (r ModList) String() string {
	return ModsToString(r)
}

func (r ModList) ApplyForInterval(value fxtools.Interval) fxtools.Interval {
	for _, mod := range r {
		value = mod.ApplyForInterval(value)
	}
	return value
}

func (r ModList) Apply(skill int) int {
	for _, mod := range r {
		skill = mod.Apply(skill)
	}
	return skill
}

type Modifier interface {
	Description() string
	Apply(int) int
	SortOrder() int
	IsNonModifying() bool
	ApplyForInterval(value fxtools.Interval) fxtools.Interval
}

func (cs *CharSheet) GetDerivedStatWithModInfo(ds DerivedStat) (int, []Modifier) {
	baseValue := cs.getDerivedStatBaseValue(ds)
	return cs.getModifiedDerivedStatWithInfo(ds, baseValue)
}

func (cs *CharSheet) GetSkill(skill Skill) int {
	baseValue := cs.GetUnmodifiedSkill(skill)
	skillValue := cs.onRetrieveSkillHook(skill, baseValue)
	return skillValue
}
func (cs *CharSheet) IsSkillAtCap(skill Skill) bool {
	return cs.GetUnmodifiedSkill(skill) >= SkillCap
}
func (cs *CharSheet) GetSkillWithModInfo(skill Skill) (int, []Modifier) {
	baseValue := cs.GetUnmodifiedSkill(skill)
	return cs.getModifiedSkillWithInfo(skill, baseValue)
}

func (cs *CharSheet) GetUnmodifiedSkill(skill Skill) int {
	return cs.getSkillBase(skill) + cs.getSkillAdjustment(skill) + cs.getTagSkillBonus(skill)
}

func (cs *CharSheet) getTagSkillBonus(skill Skill) int {
	if cs.taggedSkills[skill] {
		return 20
	}
	return 0
}

func (cs *CharSheet) getSkillAdjustment(skill Skill) int {
	if value, ok := cs.skillAdjustments[skill]; ok {
		return value
	}
	return 0
}

func (cs *CharSheet) getDerivedStatAdjustment(ds DerivedStat) int {
	if value, ok := cs.derivedStatAdjustments[ds]; ok {
		return value
	}
	return 0
}

func (cs *CharSheet) SetSkillAdjustment(skill Skill, value int) {
	cs.skillAdjustments[skill] = value
}

func (cs *CharSheet) GetDerivedStat(ds DerivedStat) int {
	baseValue := cs.getDerivedStatBaseValue(ds) + cs.getDerivedStatAdjustment(ds)
	derivedStatValue := cs.onRetrieveDerivedStatHook(ds, baseValue)
	return derivedStatValue
}

func (cs *CharSheet) getDerivedStatBaseValue(ds DerivedStat) int {
	if expr, exists := derivedBaseValues[ds]; exists {
		result, _ := expr.Evaluate(nil)
		return int(result.(float64))
	}
	return 0
}

func (cs *CharSheet) HealAPAndHPCompletely() {
	cs.hitPointsCurrent = cs.GetDerivedStat(HitPoints)
	cs.actionPointsCurrent = cs.GetDerivedStat(ActionPoints)

	cs.onDerivedStatChanged(HitPoints)
	cs.onDerivedStatChanged(ActionPoints)
}

func (cs *CharSheet) TakeRawDamage(damage int) {
	cs.hitPointsCurrent -= damage
	cs.onDerivedStatChanged(HitPoints)
}

func (cs *CharSheet) IsAlive() bool {
	return cs.hitPointsCurrent > 0
}

func (cs *CharSheet) getModifiedDerivedStatWithInfo(ds DerivedStat, value int) (int, []Modifier) {
	if cs.getDerivedStatMods != nil {
		mods := cs.getDerivedStatMods(ds)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value, mods
	}
	return value, nil
}
func (cs *CharSheet) onRetrieveDerivedStatHook(ds DerivedStat, value int) int {
	if cs.getDerivedStatMods != nil {
		mods := cs.getDerivedStatMods(ds)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value
	}
	return value
}

func (cs *CharSheet) onRetrieveSkillHook(skill Skill, value int) int {
	if cs.getSkillMods != nil {
		mods := cs.getSkillMods(skill)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value
	}
	return value
}

func (cs *CharSheet) getModifiedSkillWithInfo(skill Skill, value int) (int, []Modifier) {
	if cs.getSkillMods != nil {
		mods := cs.getSkillMods(skill)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value, mods
	}
	return value, nil
}

func (cs *CharSheet) GetHitPointsMax() int {
	return cs.GetDerivedStat(HitPoints)
}

func (cs *CharSheet) GetHitPoints() int {
	return min(cs.hitPointsCurrent, cs.GetHitPointsMax())
}

func (cs *CharSheet) Heal(amount int) {
	cs.hitPointsCurrent = min(cs.hitPointsCurrent+amount, cs.GetHitPointsMax())
}

func (cs *CharSheet) GetActionPointsMax() int {
	return cs.GetDerivedStat(ActionPoints)
}

func (cs *CharSheet) GetActionPoints() int {
	return min(cs.actionPointsCurrent, cs.GetActionPointsMax())
}

func (cs *CharSheet) LooseActionPoints(amount int) {
	cs.actionPointsCurrent = max(0, cs.actionPointsCurrent-amount)
	cs.onDerivedStatChanged(ActionPoints)
}

func (cs *CharSheet) AddSkillPoints(amount int) {
	cs.availableSkillPoints += amount
}

func (cs *CharSheet) SetOnDerivedStatChangeHandler(changed func(DerivedStat)) {
	cs.onDerivedStatChangedHandler = changed
}
func (cs *CharSheet) SetOnSkillChangeHandler(changed func(Skill)) {
	cs.onSkillChangedHandler = changed
}
func (cs *CharSheet) onDerivedStatChanged(ds DerivedStat) {
	if cs.onDerivedStatChangedHandler != nil {
		cs.onDerivedStatChangedHandler(ds)
	}
}
func (cs *CharSheet) onSkillChanged(skill Skill) {
	if cs.onSkillChangedHandler != nil {
		cs.onSkillChangedHandler(skill)
	}
}

func (cs *CharSheet) SetDerivedStatAbsoluteValue(stat DerivedStat, value int) {
	currentValue := cs.GetDerivedStat(stat)
	if currentValue == value {
		return
	}
	delta := value - currentValue
	cs.derivedStatAdjustments[stat] = delta
}

func (cs *CharSheet) SetSkillAbsoluteValue(skill Skill, value int) {
	currentValue := cs.GetSkill(skill)
	if currentValue == value {
		return
	}
	delta := value - currentValue
	cs.skillAdjustments[skill] = delta
}

func (cs *CharSheet) Kill() {
	cs.hitPointsCurrent = 0
	cs.onDerivedStatChanged(HitPoints)
}

func (cs *CharSheet) IsSkillHigherOrEqual(skill Skill, difficulty int) bool {
	return cs.GetSkill(skill) >= difficulty
}

func (cs *CharSheet) SkillRollVsDiff(skill Skill, diff Difficulty) CheckResult {
	critChance := cs.GetDerivedStat(CriticalChance)
	baseSkill := cs.GetSkill(skill)
	skillModifiedByDiff := applyDifficulty(baseSkill, diff)
	cappedSuccessChange := max(0, min(SuccessChanceCap, skillModifiedByDiff))
	return SuccessRoll(Percentage(cappedSuccessChange), Percentage(critChance))
}

func (cs *CharSheet) SkillRoll(skill Skill, modifiers int) CheckResult {
	critChance := cs.GetDerivedStat(CriticalChance)
	cappedSuccessChange := max(0, min(SuccessChanceCap, cs.GetSkill(skill)+modifiers))
	return SuccessRoll(Percentage(cappedSuccessChange), Percentage(critChance))
}

func (cs *CharSheet) GetHitPointsString() string {
	return fmt.Sprintf("%d/%d", cs.GetHitPoints(), cs.GetHitPointsMax())
}

func (cs *CharSheet) ToRecord() recfile.Record {
	record := recfile.Record{
		recfile.Field{Name: "AvailableSkillPoints", Value: recfile.IntStr(cs.availableSkillPoints)},
		recfile.Field{Name: "AvailablePerks", Value: recfile.IntStr(cs.availablePerks)},
		recfile.Field{Name: "HitPoints", Value: recfile.IntStr(cs.GetHitPoints())},
		recfile.Field{Name: "ActionPoints", Value: recfile.IntStr(cs.GetActionPoints())},
	}
	// add skills
	for skillNo := 0; skillNo < SkillCount(); skillNo++ {
		skill := Skill(skillNo)
		record = append(record, recfile.Field{Name: skill.ToAdjustmentString(), Value: recfile.IntStr(cs.getSkillAdjustment(skill))})
	}

	return record
}

func (cs *CharSheet) IsTagSkill(skill Skill) bool {
	return cs.taggedSkills[skill]
}

func (cs *CharSheet) HasSkillPointsToSpend() bool {
	return cs.availableSkillPoints > 0
}

func (cs *CharSheet) TagSkill(skill Skill) {
	if cs.GetTagSkillCount() >= 3 {
		return
	}
	cs.taggedSkills[skill] = true
}

func (cs *CharSheet) UntagSkill(skill Skill) {
	delete(cs.taggedSkills, skill)
}

func (cs *CharSheet) GetTagSkillCount() int {
	return len(cs.taggedSkills)
}

func (cs *CharSheet) GetSkillPointsToSpend() int {
	return cs.availableSkillPoints
}

func (cs *CharSheet) ResetTagSkills() {
	cs.taggedSkills = make(map[Skill]bool)
}

func (cs *CharSheet) SpendSkillPoints(skill Skill, points int) {
	if cs.availableSkillPoints < points {
		return
	}
	if cs.IsSkillAtCap(skill) {
		return
	}
	cs.availableSkillPoints -= points
	isTagSkill := cs.IsTagSkill(skill)
	if isTagSkill {
		points = points * 2
	}
	cs.skillAdjustments[skill] += points
	cs.onSkillChanged(skill)
}

func (cs *CharSheet) SetSkillModifierHandler(handler func(skill Skill) []Modifier) {
	cs.getSkillMods = handler
}

func (cs *CharSheet) SetDerivedStatModifierHandler(handler func(ds DerivedStat) []Modifier) {
	cs.getDerivedStatMods = handler
}

func (cs *CharSheet) AddSkillPointsTo(skill Skill, increase int) {
	if cs.IsSkillAtCap(skill) {
		return
	}
	cs.skillAdjustments[skill] = cs.skillAdjustments[skill] + increase
}

func (cs *CharSheet) GetPerks() []PerkLevel {
	return cs.perks
}

func (cs *CharSheet) AddPerk(perkID Perk) {
	if cs.availablePerks <= 0 {
		return
	}
	cs.availablePerks--

	for i, perk := range cs.perks {
		if perk.Perk == perkID {
			cs.perks[i] = PerkLevel{Perk: perkID, Level: perk.Level + 1}
			return
		}
	}
	cs.perks = append(cs.perks, PerkLevel{Perk: perkID, Level: 1})
}

func (cs *CharSheet) HasPerk(perkID Perk) bool {
	for _, perk := range cs.perks {
		if perk.Perk == perkID {
			return true
		}
	}
	return false
}

func (cs *CharSheet) AddPerkPoints(count int) {
	cs.availablePerks += count
}

func (cs *CharSheet) CanChooseNewPerk() bool {
	return cs.availablePerks > 0
}

func (cs *CharSheet) GetPerkLevel(perkID Perk) int {
	for _, perk := range cs.perks {
		if perk.Perk == perkID {
			return perk.Level
		}
	}
	return 0
}

func (cs *CharSheet) MeetsRequirements(requirements CharacterRequirement) bool {
	for skill, neededValue := range requirements.Skills {
		if cs.GetSkill(skill) < neededValue {
			return false
		}
	}
	for derivedStat, neededValue := range requirements.DerivedStats {
		if cs.GetDerivedStat(derivedStat) < neededValue {
			return false
		}
	}
	for perk, neededLevel := range requirements.Perks {
		if cs.GetPerkLevel(perk) < neededLevel {
			return false
		}
	}
	return true
}

func (cs *CharSheet) SetGodLike() {
	for skill := 0; skill < SkillCount(); skill++ {
		cs.skillAdjustments[Skill(skill)] = SkillCap
	}
	cs.derivedStatAdjustments[HitPoints] = 999
	cs.derivedStatAdjustments[ActionPoints] = 20
	cs.availableSkillPoints = 0
}

type Difficulty int

const (
	Trivial Difficulty = iota
	VeryEasy
	Easy
	Medium
	Hard
	VeryHard
	SuperHuman
)

func (d Difficulty) GetRollModifier() int {
	switch d {
	case VeryEasy:
		return 10
	case Easy:
		return 5
	case Medium:
		return 0
	case Hard:
		return -20
	case VeryHard:
		return -40
	}
	return 0
}

func (d Difficulty) LockReductionFactor() float64 {
	switch d {
	case VeryEasy:
		return 1.5
	case Easy:
		return 1
	case Medium:
		return 0.5
	case Hard:
		return 0.35
	case VeryHard:
		return 0.1
	}
	return 1
}

func (d Difficulty) EPicksNeeded() int {
	switch d {
	case VeryEasy:
		return 10
	case Easy:
		return 20
	case Medium:
		return 30
	case Hard:
		return 40
	case VeryHard:
		return 50
	}
	return 1
}

func (d Difficulty) String() string {
	switch d {
	case Trivial:
		return "Trivial"
	case VeryEasy:
		return "Very Easy"
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	case VeryHard:
		return "Very Hard"
	case SuperHuman:
		return "Super Human"
	}
	return "Unknown"
}

func DifficultyFromString(diff string) Difficulty {
	switch strings.ToLower(diff) {
	case "trivial":
		return Trivial
	case "veryeasy":
		return VeryEasy
	case "easy":
		return Easy
	case "medium":
		return Medium
	case "hard":
		return Hard
	case "veryhard":
		return VeryHard
	case "superhuman":
		return SuperHuman
	}
	panic("invalid difficulty")
}

var skillDiffs = map[Difficulty]*govaluate.EvaluableExpression{
	Trivial:    panicHandle(govaluate.NewEvaluableExpressionWithFunctions("skill+90", standardFunctions())),
	VeryEasy:   panicHandle(govaluate.NewEvaluableExpressionWithFunctions("skill+60", standardFunctions())),
	Easy:       panicHandle(govaluate.NewEvaluableExpressionWithFunctions("skill+30", standardFunctions())),
	Medium:     panicHandle(govaluate.NewEvaluableExpressionWithFunctions("skill", standardFunctions())),
	Hard:       panicHandle(govaluate.NewEvaluableExpressionWithFunctions("skill-30", standardFunctions())),
	VeryHard:   panicHandle(govaluate.NewEvaluableExpressionWithFunctions("skill-60", standardFunctions())),
	SuperHuman: panicHandle(govaluate.NewEvaluableExpressionWithFunctions("skill-90", standardFunctions())),
}

func applyDifficulty(skill int, diff Difficulty) int {
	if expr, exists := skillDiffs[diff]; exists {
		result, _ := expr.Evaluate(map[string]interface{}{"skill": skill})
		return int(result.(float64))
	}
	panic("invalid difficulty")
}

func panicHandle(expr *govaluate.EvaluableExpression, error error) *govaluate.EvaluableExpression {
	if error != nil {
		panic(error)
	}
	return expr
}

func (cs *CharSheet) GobEncode() ([]byte, error) {
	buffer := &bytes.Buffer{}
	gobber := gob.NewEncoder(buffer)

	if err := gobber.Encode(cs.availableSkillPoints); err != nil {
		return nil, err
	}
	if err := gobber.Encode(cs.availablePerks); err != nil {
		return nil, err
	}

	if err := gobber.Encode(cs.perks); err != nil {
		return nil, err
	}
	if err := gobber.Encode(cs.derivedStatAdjustments); err != nil {
		return nil, err
	}
	if err := gobber.Encode(cs.skillAdjustments); err != nil {
		return nil, err
	}
	if err := gobber.Encode(cs.taggedSkills); err != nil {
		return nil, err
	}
	if err := gobber.Encode(cs.hitPointsCurrent); err != nil {
		return nil, err
	}
	if err := gobber.Encode(cs.actionPointsCurrent); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (cs *CharSheet) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobber := gob.NewDecoder(buffer)

	if err := gobber.Decode(&cs.availableSkillPoints); err != nil {
		return err
	}
	if err := gobber.Decode(&cs.availablePerks); err != nil {
		return err
	}

	if err := gobber.Decode(&cs.perks); err != nil {
		return err
	}
	if err := gobber.Decode(&cs.derivedStatAdjustments); err != nil {
		return err
	}
	if err := gobber.Decode(&cs.skillAdjustments); err != nil {
		return err
	}
	if err := gobber.Decode(&cs.taggedSkills); err != nil {
		return err
	}
	if err := gobber.Decode(&cs.hitPointsCurrent); err != nil {
		return err
	}
	if err := gobber.Decode(&cs.actionPointsCurrent); err != nil {
		return err
	}
	return nil
}

func (cs *CharSheet) NeedsHealing() bool {
	return cs.GetHitPoints() < cs.GetHitPointsMax()
}
