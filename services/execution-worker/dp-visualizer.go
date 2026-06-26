package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// TestCaseDef represents a single test case definition.
type TestCaseDef struct {
	ID        int         `json:"id"`
	Input     interface{} `json:"input"`
	Expected  interface{} `json:"expected"`
	Weight    int         `json:"weight,omitempty"`
	IsHidden  bool        `json:"is_hidden,omitempty"`
}

// ProblemDefinition represents a problem with test cases for DP visualization.
type DPProblemDefinition struct {
	ID            int           `json:"id"`
	Title         string        `json:"title"`
	Difficulty    string        `json:"difficulty"`
	Type          string        `json:"type"`
	MaxScore      int           `json:"max_score"`
	TimeLimitMs   int           `json:"time_limit_ms"`
	MemoryLimitMb int           `json:"memory_limit_mb"`
	FunctionSig   string        `json:"function_sig,omitempty"`
	FunctionName  string        `json:"function_name,omitempty"`
	TestCases     []TestCaseDef `json:"test_cases"`
}

// DPTableResult holds the complete DP table visualization data.
type DPTableResult struct {
	ProblemID       int           `json:"problem_id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	Table           [][]DPCell    `json:"table"`
	Steps           []DPStep      `json:"steps"`
	Dimensions      DPDimensions  `json:"dimensions"`
	Approach        string        `json:"approach"`
	TimeComplexity  string        `json:"time_complexity"`
	SpaceComplexity string        `json:"space_complexity"`
	MemoStack       []MemoFrame   `json:"memo_stack,omitempty"`
	BacktrackPath   []DPCoord     `json:"backtrack_path,omitempty"`
	OptimalCells    []DPCoord     `json:"optimal_cells,omitempty"`
	StateColors     map[string]string `json:"state_colors,omitempty"`
}

// DPCell represents a single cell in the DP table.
type DPCell struct {
	Row          int         `json:"row"`
	Col          int         `json:"col"`
	Value        interface{} `json:"value"`
	Color        string      `json:"color"`
	Formula      string      `json:"formula,omitempty"`
	Dependencies []DPCoord   `json:"dependencies,omitempty"`
	IsBaseCase   bool        `json:"is_base_case"`
	IsResult     bool        `json:"is_result"`
	IsBacktrack  bool        `json:"is_backtrack,omitempty"`
	Explanation  string      `json:"explanation,omitempty"`
}

// DPCoord represents a coordinate in the DP table.
type DPCoord struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// DPStep represents a single animation step.
type DPStep struct {
	StepNumber    int         `json:"step_number"`
	Description   string      `json:"description"`
	CellsModified []DPCell    `json:"cells_modified"`
	ActiveCells   []DPCoord   `json:"active_cells"`
	HighlightCells []DPCoord  `json:"highlight_cells"`
	CurrentValue  interface{} `json:"current_value,omitempty"`
	Formula       string      `json:"formula,omitempty"`
	MemoState     []MemoFrame `json:"memo_state,omitempty"`
	ComplexityNow string      `json:"complexity_now,omitempty"`
}

// DPDimensions holds the dimensions of the DP table.
type DPDimensions struct {
	Rows      int      `json:"rows"`
	Cols      int      `json:"cols"`
	RowLabels []string `json:"row_labels"`
	ColLabels []string `json:"col_labels"`
}

// DPProblemInput holds the input for generating a DP visualization.
type DPProblemInput struct {
	ProblemID int         `json:"problem_id"`
	Input     interface{} `json:"input"`
	Type      string      `json:"type"`
}

// MemoFrame represents a single frame in the memoization call stack.
type MemoFrame struct {
	Function string      `json:"function"`
	Args     interface{} `json:"args"`
	Result   interface{} `json:"result,omitempty"`
	Depth    int         `json:"depth"`
	Cached   bool        `json:"cached"`
}

// DPVisualizationConfig holds configuration for the visualization.
type DPVisualizationConfig struct {
	ShowBacktrack   bool `json:"show_backtrack"`
	ShowMemoStack   bool `json:"show_memo_stack"`
	ShowComplexity  bool `json:"show_complexity"`
	AnimatePerCell  bool `json:"animate_per_cell"`
	ShowDualView    bool `json:"show_dual_view"` // Memoization + Tabulation
}

// ============================================================================
// Color coding for state transitions
// ============================================================================

var stateColors = map[string]string{
	"empty":         "#f0f0f0",
	"base_case":     "#dbeafe", // Blue
	"filling":       "#fef3c7", // Yellow
	"filled":        "#d1fae5", // Green
	"highlight":     "#fce7f3", // Pink
	"result":        "#ede9fe", // Purple
	"backtrack":     "#f97316", // Orange
	"optimal_base":  "#06b6d4", // Cyan
	"comparing":     "#f59e0b", // Amber
	"dependency":    "#a7f3d0", // Light green
	"cached":        "#93c5fd", // Light blue
}

// generateDPTable generates a step-by-step DP table visualization for the given problem.
// It supports common DP patterns including:
// - 0/1 Knapsack
// - Longest Common Subsequence (LCS)
// - Fibonacci sequence
// - Edit Distance
// - Coin Change
// - Longest Increasing Subsequence (LIS)
func generateDPTable(problem DPProblemDefinition, input interface{}, config ...DPVisualizationConfig) DPTableResult {
	cfg := DPVisualizationConfig{
		ShowBacktrack:  true,
		ShowMemoStack:  true,
		ShowComplexity: true,
		AnimatePerCell: true,
		ShowDualView:   false,
	}
	if len(config) > 0 {
		cfg = config[0]
	}

	result := DPTableResult{
		StateColors: stateColors,
	}

	switch strings.ToLower(problem.Type) {
	case "knapsack", "0/1_knapsack":
		r := generateKnapsackDP(problem, input)
		result = r
	case "lcs", "longest_common_subsequence":
		r := generateLCSDP(problem, input)
		result = r
		if cfg.ShowBacktrack {
			result.BacktrackPath = computeLCSBacktrack(input)
		}
	case "fibonacci":
		r := generateFibonacciDP(problem, input)
		result = r
		if cfg.ShowMemoStack {
			result.MemoStack = []MemoFrame{}
		}
	case "edit_distance":
		r := generateEditDistanceDP(problem, input)
		result = r
		if cfg.ShowBacktrack {
			result.BacktrackPath = computeEditDistBacktrack(input)
		}
	case "coin_change":
		r := generateCoinChangeDP(problem, input)
		result = r
	case "lis", "longest_increasing_subsequence":
		r := generateLISDP(problem, input)
		result = r
		if cfg.ShowBacktrack {
			result.BacktrackPath = computeLISBacktrack(input)
		}
	default:
		r := generateGenericDP(problem, input)
		result = r
	}

	if cfg.ShowComplexity {
		result.TimeComplexity = computeTimeComplexity(problem.Type, input)
		result.SpaceComplexity = computeSpaceComplexity(problem.Type, input)
	}

	result.StateColors = stateColors
	return result
}

// ============================================================================
// Fibonacci (1D Table)
// ============================================================================

func generateFibonacciDP(problem DPProblemDefinition, input interface{}) DPTableResult {
	n, ok := toInt(input)
	if !ok || n <= 0 {
		return createErrorResult(problem, "invalid input for fibonacci")
	}

	if n > 100 {
		n = 100
	}

	// Generate both memoization and tabulation approaches
	memo := make(map[int]int)
	memoSteps := make([]DPStep, 0)
	memoStack := make([]MemoFrame, 0)
	stepNum := 0
	callDepth := 0

	var fibMemo func(int) int
	fibMemo = func(k int) int {
		callDepth++
		if k <= 1 {
			frame := MemoFrame{
				Function: "fib",
				Args:     k,
				Result:   k,
				Depth:    callDepth,
				Cached:   false,
			}
			memoStack = append(memoStack, frame)
			callDepth--
			return k
		}
		if v, ok := memo[k]; ok {
			frame := MemoFrame{
				Function: "fib",
				Args:     k,
				Result:   v,
				Depth:    callDepth,
				Cached:   true,
			}
			memoStack = append(memoStack, frame)

			stepNum++
			memoSteps = append(memoSteps, DPStep{
				StepNumber:  stepNum,
				Description: fmt.Sprintf("fib(%d) found in memo = %d (cache hit)", k, v),
				ActiveCells: []DPCoord{{Row: 0, Col: k}},
				CurrentValue: v,
				MemoState:   append([]MemoFrame{}, memoStack...),
				ComplexityNow: fmt.Sprintf("O(1) lookup — cached"),
			})
			callDepth--
			return v
		}

		stepNum++
		left := fibMemo(k-1)
		right := fibMemo(k-2)
		val := left + right
		memo[k] = val

		memoStack = append(memoStack, MemoFrame{
			Function: "fib",
			Args:     k,
			Result:   val,
			Depth:    callDepth,
			Cached:   false,
		})

		memoSteps = append(memoSteps, DPStep{
			StepNumber:  stepNum,
			Description: fmt.Sprintf("fib(%d) = fib(%d) + fib(%d) = %d + %d = %d", k, k-1, k-2, left, right, val),
			CellsModified: []DPCell{{
				Row:         0,
				Col:         k,
				Value:       val,
				Color:       "filled",
				Formula:     fmt.Sprintf("fib(%d) = fib(%d) + fib(%d)", k, k-1, k-2),
				Dependencies: []DPCoord{{Row: 0, Col: k - 1}, {Row: 0, Col: k - 2}},
				Explanation:  fmt.Sprintf("fib(%d) = %d + %d = %d", k, left, right, val),
			}},
			ActiveCells:    []DPCoord{{Row: 0, Col: k}},
			HighlightCells: []DPCoord{{Row: 0, Col: k - 1}, {Row: 0, Col: k - 2}},
			CurrentValue:   val,
			Formula:        fmt.Sprintf("%d + %d", left, right),
			MemoState:      append([]MemoFrame{}, memoStack...),
			ComplexityNow:  fmt.Sprintf("O(%d) calls so far", stepNum),
		})

		callDepth--
		return val
	}

	result := fibMemo(n)

	// Build table (1D for fibonacci)
	table := make([][]DPCell, 1)
	table[0] = make([]DPCell, n+1)
	for i := 0; i <= n; i++ {
		color := "filled"
		if i <= 1 {
			color = "base_case"
		}
		if i == n {
			color = "result"
		}
		val, ok := memo[i]
		if !ok {
			val = 0
		}
		if i <= 1 {
			val = i
		}
		table[0][i] = DPCell{
			Row:      0,
			Col:      i,
			Value:    val,
			Color:    color,
			IsResult: i == n,
		}
	}

	colLabels := make([]string, n+1)
	for i := 0; i <= n; i++ {
		colLabels[i] = fmt.Sprintf("%d", i)
	}

	return DPTableResult{
		ProblemID:       problem.ID,
		Title:           problem.Title,
		Description:     fmt.Sprintf("Fibonacci(%d) = %d — Computing bottom-up with memoization", n, result),
		Table:           table,
		Steps:           memoSteps,
		Approach:        "memoization",
		TimeComplexity:  "O(n)",
		SpaceComplexity: "O(n)",
		MemoStack:       memoStack,
		Dimensions: DPDimensions{
			Rows:      1,
			Cols:      n + 1,
			RowLabels: []string{"fib(n)"},
			ColLabels: colLabels,
		},
	}
}

// ============================================================================
// Knapsack (2D Table)
// ============================================================================

func generateKnapsackDP(problem DPProblemDefinition, input interface{}) DPTableResult {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return createErrorResult(problem, "invalid input format for knapsack")
	}

	weights, _ := toIntSlice(inputMap["weights"])
	values, _ := toIntSlice(inputMap["values"])
	capacity, _ := toInt(inputMap["capacity"])

	n := len(weights)
	if n == 0 || capacity <= 0 {
		return createErrorResult(problem, "empty input")
	}

	dp := make([][]int, n+1)
	choice := make([][]bool, n+1) // track which items are selected
	for i := range dp {
		dp[i] = make([]int, capacity+1)
		choice[i] = make([]bool, capacity+1)
	}

	steps := make([]DPStep, 0)
	stepNum := 0

	// Base case
	stepNum++
	baseCells := make([]DPCell, 0)
	for w := 0; w <= capacity; w++ {
		baseCells = append(baseCells, DPCell{
			Row:        0,
			Col:        w,
			Value:      0,
			Color:      "base_case",
			IsBaseCase: true,
			Explanation: "Base case: 0 items selected",
		})
	}
	steps = append(steps, DPStep{
		StepNumber:    stepNum,
		Description:   "Initialize base case: dp[0][w] = 0 for all capacities",
		CellsModified: baseCells,
		Formula:       "dp[0][w] = 0",
		ComplexityNow: "O(n × W) — iterating items × capacities",
	})

	// Fill the table with optimal substructure highlighting
	for i := 1; i <= n; i++ {
		for w := 0; w <= capacity; w++ {
			stepNum++
			deps := []DPCoord{{Row: i - 1, Col: w}}

			if weights[i-1] > w {
				dp[i][w] = dp[i-1][w]
				choice[i][w] = false
				steps = append(steps, DPStep{
					StepNumber:  stepNum,
					Description: fmt.Sprintf("Item %d (weight=%d) exceeds capacity %d, skip", i, weights[i-1], w),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          w,
						Value:        dp[i][w],
						Color:        "filled",
						Formula:      fmt.Sprintf("dp[%d][%d] = dp[%d][%d] = %d", i, w, i-1, w, dp[i][w]),
						Dependencies: deps,
						Explanation:  fmt.Sprintf("weight[%d]=%d > %d, cannot include", i, weights[i-1], w),
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: w}},
					HighlightCells: deps,
					CurrentValue:   dp[i][w],
					Formula:        fmt.Sprintf("dp[%d][%d] = dp[%d][%d]", i, w, i-1, w),
					ComplexityNow:  fmt.Sprintf("Step %d/%d — O(n×W)", stepNum, (n+1)*(capacity+1)),
				})
			} else {
				include := dp[i-1][w-weights[i-1]] + values[i-1]
				exclude := dp[i-1][w]
				deps = append(deps, DPCoord{Row: i - 1, Col: w - weights[i-1]})

				optSelected := include >= exclude
				dp[i][w] = max(include, exclude)
				choice[i][w] = optSelected

				cellColor := "filled"
				if include > exclude {
					cellColor = "highlight"
				}

				steps = append(steps, DPStep{
					StepNumber:  stepNum,
					Description: fmt.Sprintf("Item %d: max(include=%d, exclude=%d) = %d", i, include, exclude, dp[i][w]),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          w,
						Value:        dp[i][w],
						Color:        cellColor,
						Formula:      fmt.Sprintf("max(dp[%d][%d]+%d, dp[%d][%d])", i-1, w-weights[i-1], values[i-1], i-1, w),
						Dependencies: deps,
						Explanation:  fmt.Sprintf("Include: %d, Exclude: %d → choose %s", include, exclude, map[bool]string{true: "include", false: "exclude"}[optSelected]),
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: w}},
					HighlightCells: deps,
					CurrentValue:   dp[i][w],
					Formula:        fmt.Sprintf("max(%d, %d)", exclude, include),
					ComplexityNow:  fmt.Sprintf("Optimal substructure: dp[%d][%d] depends on dp[%d][%d] and dp[%d][%d]", i, w, i-1, w, i-1, w-weights[i-1]),
				})
			}
		}
	}

	// Backtrack to find optimal subset
	backtrackPath := make([]DPCoord, 0)
	items := make([]int, 0)
	remaining := capacity
	for i := n; i > 0; i-- {
		if choice[i][remaining] {
			items = append(items, i-1)
			backtrackPath = append(backtrackPath, DPCoord{Row: i, Col: remaining})
			remaining -= weights[i-1]
		}
	}

	// Build table
	table := make([][]DPCell, n+1)
	for i := 0; i <= n; i++ {
		table[i] = make([]DPCell, capacity+1)
		for w := 0; w <= capacity; w++ {
			color := "filled"
			if i == 0 || w == 0 {
				color = "base_case"
			}
			if i == n && w == capacity {
				color = "result"
			}
			// Check if this cell is in the backtrack path
			isBacktrack := false
			for _, bp := range backtrackPath {
				if bp.Row == i && bp.Col == w {
					isBacktrack = true
					color = "backtrack"
					break
				}
			}
			table[i][w] = DPCell{
				Row:         i,
				Col:         w,
				Value:       dp[i][w],
				Color:       color,
				IsResult:    i == n && w == capacity,
				IsBacktrack: isBacktrack,
			}
		}
	}

	rowLabels := make([]string, n+1)
	rowLabels[0] = "∅ (no items)"
	for i := 1; i <= n; i++ {
		rowLabels[i] = fmt.Sprintf("Item %d (w=%d,v=%d)", i, weights[i-1], values[i-1])
	}

	colLabels := make([]string, capacity+1)
	for w := 0; w <= capacity; w++ {
		colLabels[w] = fmt.Sprintf("W=%d", w)
	}

	// Mark optimal cells
	optimalCells := make([]DPCoord, 0)
	for _, bp := range backtrackPath {
		optimalCells = append(optimalCells, bp)
	}
	optimalCells = append(optimalCells, DPCoord{Row: n, Col: capacity})

	return DPTableResult{
		ProblemID:       problem.ID,
		Title:           problem.Title,
		Description:     fmt.Sprintf("0/1 Knapsack: Maximize value ≤ %d capacity. Selected items: %v (total weight: %d)", capacity, items, capacity-remaining),
		Table:           table,
		Steps:           steps,
		Approach:        "tabulation",
		TimeComplexity:  fmt.Sprintf("O(n×W) = O(%d×%d)", n, capacity),
		SpaceComplexity: fmt.Sprintf("O(n×W) = O(%d×%d)", n, capacity),
		BacktrackPath:   backtrackPath,
		OptimalCells:    optimalCells,
		Dimensions: DPDimensions{
			Rows:      n + 1,
			Cols:      capacity + 1,
			RowLabels: rowLabels,
			ColLabels: colLabels,
		},
	}
}

// ============================================================================
// LCS (2D Table with Backtracking)
// ============================================================================

func generateLCSDP(problem DPProblemDefinition, input interface{}) DPTableResult {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return createErrorResult(problem, "invalid input format for LCS")
	}

	str1, _ := inputMap["str1"].(string)
	str2, _ := inputMap["str2"].(string)

	m := len(str1)
	n := len(str2)

	if m == 0 || n == 0 {
		return createErrorResult(problem, "empty input strings")
	}

	dp := make([][]int, m+1)
	arrow := make([][]string, m+1) // "diag", "up", "left"
	for i := range dp {
		dp[i] = make([]int, n+1)
		arrow[i] = make([]string, n+1)
	}

	steps := make([]DPStep, 0)
	stepNum := 0

	// Base case
	stepNum++
	baseCells := make([]DPCell, 0)
	for j := 0; j <= n; j++ {
		baseCells = append(baseCells, DPCell{
			Row:        0,
			Col:        j,
			Value:      0,
			Color:      "base_case",
			IsBaseCase: true,
		})
	}
	steps = append(steps, DPStep{
		StepNumber:    stepNum,
		Description:   "Initialize base case: dp[0][j] = 0 — no characters from first string",
		CellsModified: baseCells,
		ComplexityNow: "O(m × n) DP table initialization",
	})

	// Fill table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			stepNum++

			if str1[i-1] == str2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				arrow[i][j] = "diag"
				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("✅ Match found: '%c' == '%c', extend LCS by 1", str1[i-1], str2[j-1]),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          j,
						Value:        dp[i][j],
						Color:        "highlight",
						Formula:      fmt.Sprintf("dp[%d][%d] = dp[%d][%d] + 1", i, j, i-1, j-1),
						Dependencies: []DPCoord{{Row: i - 1, Col: j - 1}},
						Explanation:  fmt.Sprintf("Characters '%c' match, LCS increases", str1[i-1]),
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: j}},
					HighlightCells: []DPCoord{{Row: i - 1, Col: j - 1}},
					CurrentValue:   dp[i][j],
					Formula:        fmt.Sprintf("dp[%d][%d] + 1", i-1, j-1),
				})
			} else {
				if dp[i-1][j] >= dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
					arrow[i][j] = "up"
				} else {
					dp[i][j] = dp[i][j-1]
					arrow[i][j] = "left"
				}
				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("❌ No match: '%c' != '%c', take max(%d, %d) = %d",
						str1[i-1], str2[j-1], dp[i-1][j], dp[i][j-1], dp[i][j]),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          j,
						Value:        dp[i][j],
						Color:        "filled",
						Formula:      fmt.Sprintf("max(dp[%d][%d], dp[%d][%d])", i-1, j, i, j-1),
						Dependencies: []DPCoord{{Row: i - 1, Col: j}, {Row: i, Col: j - 1}},
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: j}},
					HighlightCells: []DPCoord{{Row: i - 1, Col: j}, {Row: i, Col: j - 1}},
					CurrentValue:   dp[i][j],
					Formula:        fmt.Sprintf("max(%d, %d)", dp[i-1][j], dp[i][j-1]),
				})
			}
		}
	}

	// Build backtrack path
	backtrackPath := make([]DPCoord, 0)
	i, j := m, n
	for i > 0 && j > 0 {
		backtrackPath = append(backtrackPath, DPCoord{Row: i, Col: j})
		switch arrow[i][j] {
		case "diag":
			i--
			j--
		case "up":
			i--
		case "left":
			j--
		}
	}

	// Build table with backtrack markers
	table := make([][]DPCell, m+1)
	for i := 0; i <= m; i++ {
		table[i] = make([]DPCell, n+1)
		for j := 0; j <= n; j++ {
			color := "filled"
			if i == 0 || j == 0 {
				color = "base_case"
			}
			isBacktrack := false
			for _, bp := range backtrackPath {
				if bp.Row == i && bp.Col == j {
					isBacktrack = true
					if arrow[i][j] == "diag" {
						color = "highlight"
					} else {
						color = "backtrack"
					}
					break
				}
			}
			if i == m && j == n {
				color = "result"
			}
			table[i][j] = DPCell{
				Row:         i,
				Col:         j,
				Value:       dp[i][j],
				Color:       color,
				IsResult:    i == m && j == n,
				IsBacktrack: isBacktrack,
			}
		}
	}

	rowLabels := make([]string, m+1)
	rowLabels[0] = "∅"
	for i := 1; i <= m; i++ {
		rowLabels[i] = fmt.Sprintf("'%c'", str1[i-1])
	}

	colLabels := make([]string, n+1)
	colLabels[0] = "∅"
	for j := 1; j <= n; j++ {
		colLabels[j] = fmt.Sprintf("'%c'", str2[j-1])
	}

	return DPTableResult{
		ProblemID:       problem.ID,
		Title:           problem.Title,
		Description:     fmt.Sprintf("LCS of \"%s\" and \"%s\" = %d (%s)", str1, str2, dp[m][n], reconstructLCS(str1, str2, arrow)),
		Table:           table,
		Steps:           steps,
		Approach:        "tabulation",
		TimeComplexity:  fmt.Sprintf("O(m×n) = O(%d×%d)", m, n),
		SpaceComplexity: fmt.Sprintf("O(m×n) = O(%d×%d)", m, n),
		BacktrackPath:   backtrackPath,
		OptimalCells:    []DPCoord{{Row: m, Col: n}},
		Dimensions: DPDimensions{
			Rows:      m + 1,
			Cols:      n + 1,
			RowLabels: rowLabels,
			ColLabels: colLabels,
		},
	}
}

func computeLCSBacktrack(input interface{}) []DPCoord {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil
	}
	str1, _ := inputMap["str1"].(string)
	str2, _ := inputMap["str2"].(string)
	m, n := len(str1), len(str2)

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if str1[i-1] == str2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	path := make([]DPCoord, 0)
	i, j := m, n
	for i > 0 && j > 0 {
		path = append(path, DPCoord{Row: i, Col: j})
		if str1[i-1] == str2[j-1] {
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	return path
}

func reconstructLCS(str1, str2 string, arrow [][]string) string {
	var result strings.Builder
	i, j := len(str1), len(str2)
	for i > 0 && j > 0 {
		switch arrow[i][j] {
		case "diag":
			result.WriteByte(str1[i-1])
			i--
			j--
		case "up":
			i--
		case "left":
			j--
		}
	}
	// Reverse
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ============================================================================
// Edit Distance (2D Table)
// ============================================================================

func generateEditDistanceDP(problem DPProblemDefinition, input interface{}) DPTableResult {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return createErrorResult(problem, "invalid input format for edit distance")
	}

	str1, _ := inputMap["str1"].(string)
	str2, _ := inputMap["str2"].(string)

	m := len(str1)
	n := len(str2)

	if m == 0 || n == 0 {
		return createErrorResult(problem, "empty input strings")
	}

	dp := make([][]int, m+1)
	op := make([][]string, m+1) // track operations: "match", "insert", "delete", "replace"
	for i := range dp {
		dp[i] = make([]int, n+1)
		op[i] = make([]string, n+1)
	}

	steps := make([]DPStep, 0)
	stepNum := 0

	// Base cases
	for i := 0; i <= m; i++ {
		dp[i][0] = i
		op[i][0] = "delete"
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j
		op[0][j] = "insert"
	}

	stepNum++
	steps = append(steps, DPStep{
		StepNumber:  stepNum,
		Description: "Initialize base cases: dp[i][0] = i (delete all), dp[0][j] = j (insert all)",
		ComplexityNow: "O(m × n) edit distance DP",
	})

	// Fill table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			stepNum++

			if str1[i-1] == str2[j-1] {
				dp[i][j] = dp[i-1][j-1]
				op[i][j] = "match"
				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("Match: '%c' == '%c', cost = 0, dp[%d][%d] = dp[%d][%d] = %d",
						str1[i-1], str2[j-1], i, j, i-1, j-1, dp[i][j]),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          j,
						Value:        dp[i][j],
						Color:        "highlight",
						Formula:      fmt.Sprintf("dp[%d][%d] = dp[%d][%d] (no cost)", i, j, i-1, j-1),
						Dependencies: []DPCoord{{Row: i - 1, Col: j - 1}},
						Explanation:  fmt.Sprintf("Characters match: '%c'", str1[i-1]),
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: j}},
					HighlightCells: []DPCoord{{Row: i - 1, Col: j - 1}},
					CurrentValue:   dp[i][j],
				})
			} else {
				insertCost := dp[i][j-1] + 1
				deleteCost := dp[i-1][j] + 1
				replaceCost := dp[i-1][j-1] + 1

				minCost := min3(insertCost, deleteCost, replaceCost)
				dp[i][j] = minCost

				if minCost == replaceCost {
					op[i][j] = "replace"
				} else if minCost == deleteCost {
					op[i][j] = "delete"
				} else {
					op[i][j] = "insert"
				}

				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("Replace '%c'→'%c': min(insert=%d, delete=%d, replace=%d) = %d",
						str1[i-1], str2[j-1], insertCost, deleteCost, replaceCost, minCost),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          j,
						Value:        minCost,
						Color:        "filled",
						Formula:      fmt.Sprintf("1 + min(dp[%d][%d], dp[%d][%d], dp[%d][%d])", i, j-1, i-1, j, i-1, j-1),
						Dependencies: []DPCoord{{Row: i, Col: j - 1}, {Row: i - 1, Col: j}, {Row: i - 1, Col: j - 1}},
						Explanation:  fmt.Sprintf("Operation: %s (cost=1) + %d", op[i][j], minCost-1),
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: j}},
					HighlightCells: []DPCoord{{Row: i, Col: j - 1}, {Row: i - 1, Col: j}, {Row: i - 1, Col: j - 1}},
					CurrentValue:   minCost,
					Formula:        fmt.Sprintf("1 + min(%d, %d, %d)", insertCost-1, deleteCost-1, replaceCost-1),
				})
			}
		}
	}

	// Build backtrack path
	backtrackPath := make([]DPCoord, 0)
	ii, jj := m, n
	for ii > 0 || jj > 0 {
		backtrackPath = append(backtrackPath, DPCoord{Row: ii, Col: jj})
		switch op[ii][jj] {
		case "match", "replace":
			ii--
			jj--
		case "delete":
			ii--
		case "insert":
			jj--
		}
	}

	// Build table
	table := make([][]DPCell, m+1)
	for i := 0; i <= m; i++ {
		table[i] = make([]DPCell, n+1)
		for j := 0; j <= n; j++ {
			color := "filled"
			if i == 0 || j == 0 {
				color = "base_case"
			}
			isBacktrack := false
			for _, bp := range backtrackPath {
				if bp.Row == i && bp.Col == j {
					isBacktrack = true
					if op[i][j] == "match" {
						color = "highlight"
					} else {
						color = "backtrack"
					}
					break
				}
			}
			if i == m && j == n {
				color = "result"
			}
			table[i][j] = DPCell{
				Row:         i,
				Col:         j,
				Value:       dp[i][j],
				Color:       color,
				IsResult:    i == m && j == n,
				IsBacktrack: isBacktrack,
			}
		}
	}

	return DPTableResult{
		ProblemID:       problem.ID,
		Title:           problem.Title,
		Description:     fmt.Sprintf("Edit Distance from \"%s\" to \"%s\" = %d operations", str1, str2, dp[m][n]),
		Table:           table,
		Steps:           steps,
		Approach:        "tabulation",
		TimeComplexity:  fmt.Sprintf("O(m×n) = O(%d×%d)", m, n),
		SpaceComplexity: fmt.Sprintf("O(m×n) = O(%d×%d)", m, n),
		BacktrackPath:   backtrackPath,
		OptimalCells:    []DPCoord{{Row: m, Col: n}},
		Dimensions: DPDimensions{
			Rows:      m + 1,
			Cols:      n + 1,
			RowLabels: makeRowLabels(str1),
			ColLabels: makeColLabels(str2),
		},
	}
}

func computeEditDistBacktrack(input interface{}) []DPCoord {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil
	}
	str1, _ := inputMap["str1"].(string)
	str2, _ := inputMap["str2"].(string)
	m, n := len(str1), len(str2)

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 0; i <= m; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if str1[i-1] == str2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min3(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
			}
		}
	}

	path := make([]DPCoord, 0)
	i, j := m, n
	for i > 0 || j > 0 {
		path = append(path, DPCoord{Row: i, Col: j})
		if i > 0 && j > 0 && str1[i-1] == str2[j-1] {
			i--
			j--
		} else if i > 0 && j > 0 && dp[i][j] == dp[i-1][j-1]+1 {
			i--
			j--
		} else if i > 0 && dp[i][j] == dp[i-1][j]+1 {
			i--
		} else {
			j--
		}
	}
	return path
}

// ============================================================================
// Coin Change (2D Table min coins)
// ============================================================================

func generateCoinChangeDP(problem DPProblemDefinition, input interface{}) DPTableResult {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return createErrorResult(problem, "invalid input format for coin change")
	}

	coins, _ := toIntSlice(inputMap["coins"])
	amount, _ := toInt(inputMap["amount"])

	if len(coins) == 0 || amount <= 0 {
		return createErrorResult(problem, "empty input")
	}

	n := len(coins)
	const INF = 1 << 30

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, amount+1)
		for j := range dp[i] {
			dp[i][j] = INF
		}
	}

	steps := make([]DPStep, 0)
	stepNum := 0

	for i := 0; i <= n; i++ {
		dp[i][0] = 0
	}

	stepNum++
	steps = append(steps, DPStep{
		StepNumber:  stepNum,
		Description: "Initialize base case: dp[i][0] = 0 — zero coins needed for zero amount",
		ComplexityNow: "O(n × amount) DP for coin change",
	})

	for i := 1; i <= n; i++ {
		for j := 1; j <= amount; j++ {
			stepNum++

			if j >= coins[i-1] {
				skip := dp[i-1][j]
				use := dp[i][j-coins[i-1]] + 1
				dp[i][j] = min(skip, use)
				optStr := "skip"
				if use < skip {
					optStr = "use"
				}
				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("Coin %d¢: min(skip=%d, use=%d+1) = %d",
						coins[i-1], skip, dp[i][j-coins[i-1]], dp[i][j]),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          j,
						Value:        dp[i][j],
						Color:        map[bool]string{true: "highlight"}[use < skip],
						Formula:      fmt.Sprintf("min(dp[%d][%d], dp[%d][%d]+1)", i-1, j, i, j-coins[i-1]),
						Dependencies: []DPCoord{{Row: i - 1, Col: j}, {Row: i, Col: j - coins[i-1]}},
						Explanation:  fmt.Sprintf("Option: %s — %s coin %d¢", optStr, map[string]string{"skip": "Skip", "use": "Use"}[optStr], coins[i-1]),
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: j}},
					HighlightCells: []DPCoord{{Row: i - 1, Col: j}, {Row: i, Col: j - coins[i-1]}},
					CurrentValue:   dp[i][j],
				})
			} else {
				dp[i][j] = dp[i-1][j]
				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("Coin %d¢ > amount %d, skip", coins[i-1], j),
					CellsModified: []DPCell{{
						Row:          i,
						Col:          j,
						Value:        dp[i][j],
						Color:        "filled",
						Formula:      fmt.Sprintf("dp[%d][%d] = dp[%d][%d]", i, j, i-1, j),
						Dependencies: []DPCoord{{Row: i - 1, Col: j}},
					}},
					ActiveCells:    []DPCoord{{Row: i, Col: j}},
					HighlightCells: []DPCoord{{Row: i - 1, Col: j}},
					CurrentValue:   dp[i][j],
				})
			}
		}
	}

	table := make([][]DPCell, n+1)
	for i := 0; i <= n; i++ {
		table[i] = make([]DPCell, amount+1)
		for j := 0; j <= amount; j++ {
			color := "filled"
			if j == 0 {
				color = "base_case"
			}
			if dp[i][j] >= INF {
				table[i][j] = DPCell{
					Row:   i,
					Col:   j,
					Value: "∞",
					Color: "empty",
				}
			} else {
				table[i][j] = DPCell{
					Row:      i,
					Col:      j,
					Value:    dp[i][j],
					Color:    color,
					IsResult: i == n && j == amount,
				}
			}
		}
	}

	result := dp[n][amount]
	if result >= INF {
		result = -1
	}

	return DPTableResult{
		ProblemID:       problem.ID,
		Title:           problem.Title,
		Description:     fmt.Sprintf("Coin Change: minimum coins to make %d¢ = %d", amount, result),
		Table:           table,
		Steps:           steps,
		Approach:        "tabulation",
		TimeComplexity:  fmt.Sprintf("O(n×amount) = O(%d×%d)", n, amount),
		SpaceComplexity: fmt.Sprintf("O(n×amount) = O(%d×%d)", n, amount),
		OptimalCells:    []DPCoord{{Row: n, Col: amount}},
		Dimensions: DPDimensions{
			Rows:      n + 1,
			Cols:      amount + 1,
			RowLabels: makeCoinRowLabels(coins),
			ColLabels: makeAmountColLabels(amount),
		},
	}
}

// ============================================================================
// Longest Increasing Subsequence (1D)
// ============================================================================

func generateLISDP(problem DPProblemDefinition, input interface{}) DPTableResult {
	var arr []int
	switch v := input.(type) {
	case map[string]interface{}:
		arr, _ = toIntSlice(v["arr"])
	default:
		arr, _ = toIntSlice(input)
	}

	if len(arr) == 0 {
		return createErrorResult(problem, "empty input array")
	}

	n := len(arr)
	dp := make([]int, n)
	prev := make([]int, n) // for backtracking
	for i := range dp {
		dp[i] = 1
		prev[i] = -1
	}

	steps := make([]DPStep, 0)
	stepNum := 0

	// Base state
	stepNum++
	baseCells := make([]DPCell, n)
	for i := range arr {
		baseCells[i] = DPCell{
			Row: 0, Col: i,
			Value:    1,
			Color:    "base_case",
			IsBaseCase: true,
			Explanation: fmt.Sprintf("Base: LIS ending at arr[%d]=%d starts at 1", i, arr[i]),
		}
	}
	steps = append(steps, DPStep{
		StepNumber:    stepNum,
		Description:   "Initialize: dp[i] = 1 for all positions (each element alone is an LIS of length 1)",
		CellsModified: baseCells,
		ComplexityNow: fmt.Sprintf("O(n²) = O(%d²) LIS algorithm", n),
	})

	// Fill dp table
	overallMax := 1
	maxIdx := 0
	for i := 0; i < n; i++ {
		for j := 0; j < i; j++ {
			stepNum++
			if arr[j] < arr[i] && dp[j]+1 > dp[i] {
				dp[i] = dp[j] + 1
				prev[i] = j
				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("arr[%d]=%d > arr[%d]=%d, extend LIS: dp[%d] = dp[%d]+1 = %d",
						i, arr[i], j, arr[j], i, j, dp[i]),
					CellsModified: []DPCell{{
						Row:          0,
						Col:          i,
						Value:        dp[i],
						Color:        "highlight",
						Formula:      fmt.Sprintf("dp[%d] = dp[%d]+1 = %d+1", i, j, dp[j]),
						Dependencies: []DPCoord{{Row: 0, Col: j}},
						Explanation:  fmt.Sprintf("Extend LIS at %d→%d: %d < %d", j, i, arr[j], arr[i]),
					}},
					ActiveCells:    []DPCoord{{Row: 0, Col: i}},
					HighlightCells: []DPCoord{{Row: 0, Col: j}},
					CurrentValue:   dp[i],
				})
			} else {
				steps = append(steps, DPStep{
					StepNumber: stepNum,
					Description: fmt.Sprintf("arr[%d]=%d not > arr[%d]=%d, skip",
						i, arr[i], j, arr[j]),
					ActiveCells:    []DPCoord{{Row: 0, Col: i}},
					HighlightCells: []DPCoord{{Row: 0, Col: j}},
					CurrentValue:   dp[i],
				})
			}
		}
		if dp[i] > overallMax {
			overallMax = dp[i]
			maxIdx = i
		}
	}

	// Build backtrack path
	backtrackPath := make([]DPCoord, 0)
	curr := maxIdx
	for curr >= 0 {
		backtrackPath = append(backtrackPath, DPCoord{Row: 0, Col: curr})
		curr = prev[curr]
	}

	// Build 1D table
	table := make([][]DPCell, 1)
	table[0] = make([]DPCell, n)
	for i := 0; i < n; i++ {
		color := "filled"
		if dp[i] == 1 {
			color = "base_case"
		}
		isBacktrack := false
		for _, bp := range backtrackPath {
			if bp.Row == 0 && bp.Col == i {
				isBacktrack = true
				color = "backtrack"
				break
			}
		}
		if i == maxIdx {
			color = "result"
		}
		table[0][i] = DPCell{
			Row:         0,
			Col:         i,
			Value:       dp[i],
			Color:       color,
			IsResult:    i == maxIdx,
			IsBacktrack: isBacktrack,
			Explanation: fmt.Sprintf("arr[%d]=%d → LIS length=%d", i, arr[i], dp[i]),
		}
	}

	colLabels := make([]string, n)
	for i := 0; i < n; i++ {
		colLabels[i] = fmt.Sprintf("arr[%d]=%d", i, arr[i])
	}

	return DPTableResult{
		ProblemID:       problem.ID,
		Title:           problem.Title,
		Description:     fmt.Sprintf("LIS length = %d, ends at index %d (value=%d)", overallMax, maxIdx, arr[maxIdx]),
		Table:           table,
		Steps:           steps,
		Approach:        "tabulation",
		TimeComplexity:  fmt.Sprintf("O(n²) = O(%d²)", n),
		SpaceComplexity: fmt.Sprintf("O(n) = O(%d)", n),
		BacktrackPath:   backtrackPath,
		OptimalCells:    []DPCoord{{Row: 0, Col: maxIdx}},
		Dimensions: DPDimensions{
			Rows:      1,
			Cols:      n,
			RowLabels: []string{"LIS Length"},
			ColLabels: colLabels,
		},
	}
}

func computeLISBacktrack(input interface{}) []DPCoord {
	var arr []int
	switch v := input.(type) {
	case map[string]interface{}:
		arr, _ = toIntSlice(v["arr"])
	default:
		arr, _ = toIntSlice(input)
	}
	if len(arr) == 0 {
		return nil
	}

	n := len(arr)
	dp := make([]int, n)
	prev := make([]int, n)
	for i := range dp {
		dp[i] = 1
		prev[i] = -1
	}

	maxLen, maxIdx := 1, 0
	for i := 0; i < n; i++ {
		for j := 0; j < i; j++ {
			if arr[j] < arr[i] && dp[j]+1 > dp[i] {
				dp[i] = dp[j] + 1
				prev[i] = j
			}
		}
		if dp[i] > maxLen {
			maxLen = dp[i]
			maxIdx = i
		}
	}

	path := make([]DPCoord, 0)
	curr := maxIdx
	for curr >= 0 {
		path = append(path, DPCoord{Row: 0, Col: curr})
		curr = prev[curr]
	}
	return path
}

// ============================================================================
// Generic DP
// ============================================================================

func generateGenericDP(problem DPProblemDefinition, input interface{}) DPTableResult {
	return DPTableResult{
		ProblemID:   problem.ID,
		Title:       problem.Title,
		Description: fmt.Sprintf("DP visualization for problem: %s", problem.Title),
		Table:       [][]DPCell{},
		Steps:       []DPStep{},
		Approach:    "tabulation",
		Dimensions:  DPDimensions{Rows: 0, Cols: 0},
	}
}

func createErrorResult(problem DPProblemDefinition, errMsg string) DPTableResult {
	return DPTableResult{
		ProblemID:   problem.ID,
		Title:       problem.Title,
		Description: errMsg,
		Table:       [][]DPCell{},
		Steps:       []DPStep{},
		Approach:    "tabulation",
		Dimensions:  DPDimensions{Rows: 0, Cols: 0},
	}
}

// ============================================================================
// Complexity Analysis
// ============================================================================

func computeTimeComplexity(dpType string, input interface{}) string {
	switch strings.ToLower(dpType) {
	case "fibonacci":
		return "O(n) — linear time, each value computed once"
	case "knapsack", "0/1_knapsack":
		if m, ok := input.(map[string]interface{}); ok {
			weights, _ := toIntSlice(m["weights"])
			capacity, _ := toInt(m["capacity"])
			return fmt.Sprintf("O(n×W) = O(%d×%d) — pseudo-polynomial", len(weights), capacity)
		}
		return "O(n×W) — pseudo-polynomial"
	case "lcs", "longest_common_subsequence":
		if m, ok := input.(map[string]interface{}); ok {
			s1, _ := m["str1"].(string)
			s2, _ := m["str2"].(string)
			return fmt.Sprintf("O(m×n) = O(%d×%d) — quadratic", len(s1), len(s2))
		}
		return "O(m×n) — quadratic"
	case "edit_distance":
		if m, ok := input.(map[string]interface{}); ok {
			s1, _ := m["str1"].(string)
			s2, _ := m["str2"].(string)
			return fmt.Sprintf("O(m×n) = O(%d×%d) — quadratic", len(s1), len(s2))
		}
		return "O(m×n) — quadratic"
	case "coin_change":
		if m, ok := input.(map[string]interface{}); ok {
			coins, _ := toIntSlice(m["coins"])
			amount, _ := toInt(m["amount"])
			return fmt.Sprintf("O(n×amount) = O(%d×%d) — pseudo-polynomial", len(coins), amount)
		}
		return "O(n×amount) — pseudo-polynomial"
	case "lis", "longest_increasing_subsequence":
		n := 0
		if m, ok := input.(map[string]interface{}); ok {
			arr, _ := toIntSlice(m["arr"])
			n = len(arr)
		}
		return fmt.Sprintf("O(n²) = O(%d²) — quadratic", n)
	default:
		return "Varies — depends on DP formulation"
	}
}

func computeSpaceComplexity(dpType string, input interface{}) string {
	switch strings.ToLower(dpType) {
	case "fibonacci":
		return "O(n) — 1D array for memoization"
	case "knapsack", "0/1_knapsack":
		if m, ok := input.(map[string]interface{}); ok {
			weights, _ := toIntSlice(m["weights"])
			capacity, _ := toInt(m["capacity"])
			return fmt.Sprintf("O(n×W) = O(%d×%d) — 2D table", len(weights), capacity)
		}
		return "O(n×W) — 2D table"
	case "lcs", "longest_common_subsequence":
		if m, ok := input.(map[string]interface{}); ok {
			s1, _ := m["str1"].(string)
			s2, _ := m["str2"].(string)
			return fmt.Sprintf("O(m×n) = O(%d×%d) — 2D table", len(s1), len(s2))
		}
		return "O(m×n) — 2D table"
	case "edit_distance":
		if m, ok := input.(map[string]interface{}); ok {
			s1, _ := m["str1"].(string)
			s2, _ := m["str2"].(string)
			return fmt.Sprintf("O(m×n) = O(%d×%d) — 2D table", len(s1), len(s2))
		}
		return "O(m×n) — 2D table"
	case "coin_change":
		if m, ok := input.(map[string]interface{}); ok {
			coins, _ := toIntSlice(m["coins"])
			amount, _ := toInt(m["amount"])
			return fmt.Sprintf("O(n×amount) = O(%d×%d) — 2D table", len(coins), amount)
		}
		return "O(n×amount) — 2D table"
	case "lis", "longest_increasing_subsequence":
		return "O(n) — 1D array"
	default:
		return "Varies"
	}
}

// ============================================================================
// Label helpers
// ============================================================================

func makeRowLabels(str string) []string {
	labels := make([]string, len(str)+1)
	labels[0] = "∅"
	for i := 1; i <= len(str); i++ {
		labels[i] = fmt.Sprintf("'%c'", str[i-1])
	}
	return labels
}

func makeColLabels(str string) []string {
	labels := make([]string, len(str)+1)
	labels[0] = "∅"
	for j := 1; j <= len(str); j++ {
		labels[j] = fmt.Sprintf("'%c'", str[j-1])
	}
	return labels
}

func makeCoinRowLabels(coins []int) []string {
	labels := make([]string, len(coins)+1)
	labels[0] = "∅ (no coins)"
	for i := 1; i <= len(coins); i++ {
		labels[i] = fmt.Sprintf("%d¢ coin", coins[i-1])
	}
	return labels
}

func makeAmountColLabels(amount int) []string {
	labels := make([]string, amount+1)
	for j := 0; j <= amount; j++ {
		labels[j] = fmt.Sprintf("%d", j)
	}
	return labels
}

// ============================================================================
// Helper functions
// ============================================================================

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	case string:
		var result int
		_, err := fmt.Sscanf(n, "%d", &result)
		return result, err == nil
	default:
		return 0, false
	}
}

func toIntSlice(v interface{}) ([]int, bool) {
	switch arr := v.(type) {
	case []interface{}:
		result := make([]int, 0, len(arr))
		for _, item := range arr {
			if n, ok := toInt(item); ok {
				result = append(result, n)
			}
		}
		return result, true
	case []int:
		return arr, true
	case string:
		var result []int
		if err := json.Unmarshal([]byte(arr), &result); err == nil {
			return result, true
		}
		return nil, false
	default:
		return nil, false
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func min3(a, b, c int) int {
	return min(min(a, b), c)
}

// Ensure math package is used
var _ = math.MaxInt32
