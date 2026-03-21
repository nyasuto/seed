package economy

import "testing"

func TestRenderEconomyStatus_NilLastTick(t *testing.T) {
	pool := NewChiPool(150.0)
	pool.Current = 45.2
	engine := &EconomyEngine{ChiPool: pool}

	got := RenderEconomyStatus(engine, nil)
	want := "[Chi: 45.2/150.0 | +0.0 -0.0 = +0.0/tick | OK]"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderEconomyStatus_WithTickResult(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		cap      float64
		tick     EconomyTickResult
		want     string
	}{
		{
			name:    "healthy economy",
			current: 45.2,
			cap:     150.0,
			tick: EconomyTickResult{
				Supply:      5.0,
				Maintenance: MaintenanceBreakdown{Total: 3.8},
				DeficitResult: DeficitResult{
					Severity: None,
				},
			},
			want: "[Chi: 45.2/150.0 | +5.0 -3.8 = +1.2/tick | OK]",
		},
		{
			name:    "mild deficit",
			current: 2.0,
			cap:     100.0,
			tick: EconomyTickResult{
				Supply:      3.0,
				Maintenance: MaintenanceBreakdown{Total: 8.0},
				DeficitResult: DeficitResult{
					Severity: Mild,
					Shortage: 3.0,
				},
			},
			want: "[Chi: 2.0/100.0 | +3.0 -8.0 = -5.0/tick | MILD]",
		},
		{
			name:    "moderate deficit",
			current: 0.0,
			cap:     100.0,
			tick: EconomyTickResult{
				Supply:      1.0,
				Maintenance: MaintenanceBreakdown{Total: 10.0},
				DeficitResult: DeficitResult{
					Severity: Moderate,
				},
			},
			want: "[Chi: 0.0/100.0 | +1.0 -10.0 = -9.0/tick | MODERATE]",
		},
		{
			name:    "severe deficit",
			current: 0.0,
			cap:     80.0,
			tick: EconomyTickResult{
				Supply:      0.5,
				Maintenance: MaintenanceBreakdown{Total: 12.0},
				DeficitResult: DeficitResult{
					Severity: Severe,
				},
			},
			want: "[Chi: 0.0/80.0 | +0.5 -12.0 = -11.5/tick | SEVERE]",
		},
		{
			name:    "zero activity",
			current: 50.0,
			cap:     50.0,
			tick: EconomyTickResult{
				Supply:      0.0,
				Maintenance: MaintenanceBreakdown{Total: 0.0},
				DeficitResult: DeficitResult{
					Severity: None,
				},
			},
			want: "[Chi: 50.0/50.0 | +0.0 -0.0 = +0.0/tick | OK]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewChiPool(tt.cap)
			pool.Current = tt.current
			engine := &EconomyEngine{ChiPool: pool}

			got := RenderEconomyStatus(engine, &tt.tick)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSeverityLabel(t *testing.T) {
	tests := []struct {
		severity DeficitSeverity
		want     string
	}{
		{None, "OK"},
		{Mild, "MILD"},
		{Moderate, "MODERATE"},
		{Severe, "SEVERE"},
	}
	for _, tt := range tests {
		got := severityLabel(tt.severity)
		if got != tt.want {
			t.Errorf("severityLabel(%d) = %q, want %q", tt.severity, got, tt.want)
		}
	}
}
