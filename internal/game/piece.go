package game

import "github.com/gdamore/tcell/v2"

// Predefined piece colors
var PieceColors = map[PieceID]tcell.Color{
	I: tcell.ColorTeal,
	O: tcell.ColorYellow,
	T: tcell.ColorPurple,
	S: tcell.ColorGreen,
	Z: tcell.ColorRed,
	J: tcell.ColorBlue,
	L: tcell.NewRGBColor(255, 165, 0), // Orange
}

// SRS offsets and block matrices.
// Each [4][4]bool is a rotation state; true = block present.
var PieceShapes = map[PieceID][4][4][4]bool{
	I: {
		// state 0
		{
			{false, false, false, false},
			{true, true, true, true},
			{false, false, false, false},
			{false, false, false, false},
		},
		// R
		{
			{false, false, true, false},
			{false, false, true, false},
			{false, false, true, false},
			{false, false, true, false},
		},
		// 2
		{
			{false, false, false, false},
			{false, false, false, false},
			{true, true, true, true},
			{false, false, false, false},
		},
		// L
		{
			{false, true, false, false},
			{false, true, false, false},
			{false, true, false, false},
			{false, true, false, false},
		},
	},
	O: {
		// All four rotation states are the same for O
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, true, true, false},
			{false, false, false, false},
		},
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, true, true, false},
			{false, false, false, false},
		},
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, true, true, false},
			{false, false, false, false},
		},
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, true, true, false},
			{false, false, false, false},
		},
	},
	T: {
		// state 0
		{
			{false, false, false, false},
			{false, true, true, true},
			{false, false, true, false},
			{false, false, false, false},
		},
		// R
		{
			{false, false, false, false},
			{false, false, true, false},
			{false, true, true, false},
			{false, false, true, false},
		},
		// 2
		{
			{false, false, false, false},
			{false, false, true, false},
			{false, true, true, true},
			{false, false, false, false},
		},
		// L
		{
			{false, false, false, false},
			{false, true, false, false},
			{false, true, true, false},
			{false, true, false, false},
		},
	},
	J: {
		// state 0
		{
			{false, false, false, false},
			{false, true, true, true},
			{false, false, false, true},
			{false, false, false, false},
		},
		// R
		{
			{false, false, false, false},
			{false, false, true, false},
			{false, false, true, false},
			{false, true, true, false},
		},
		// 2
		{
			{false, false, false, false},
			{false, true, false, false},
			{false, true, true, true},
			{false, false, false, false},
		},
		// L
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, true, false, false},
			{false, true, false, false},
		},
	},
	L: {
		// state 0
		{
			{false, false, false, false},
			{false, true, true, true},
			{false, true, false, false},
			{false, false, false, false},
		},
		// R
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, false, true, false},
			{false, false, true, false},
		},
		// 2
		{
			{false, false, false, false},
			{false, false, false, true},
			{false, true, true, true},
			{false, false, false, false},
		},
		// L
		{
			{false, false, false, false},
			{false, true, false, false},
			{false, true, false, false},
			{false, true, true, false},
		},
	},
	S: {
		// state 0
		{
			{false, false, false, false},
			{false, false, true, true},
			{false, true, true, false},
			{false, false, false, false},
		},
		// R
		{
			{false, false, false, false},
			{false, true, false, false},
			{false, true, true, false},
			{false, false, true, false},
		},
		// 2
		{
			{false, false, false, false},
			{false, false, true, true},
			{false, true, true, false},
			{false, false, false, false},
		},
		// L
		{
			{false, false, false, false},
			{false, true, false, false},
			{false, true, true, false},
			{false, false, true, false},
		},
	},
	Z: {
		// state 0
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, false, true, true},
			{false, false, false, false},
		},
		// R
		{
			{false, false, false, false},
			{false, false, true, false},
			{false, true, true, false},
			{false, true, false, false},
		},
		// 2
		{
			{false, false, false, false},
			{false, true, true, false},
			{false, false, true, true},
			{false, false, false, false},
		},
		// L
		{
			{false, false, false, false},
			{false, false, true, false},
			{false, true, true, false},
			{false, true, false, false},
		},
	},
}

// ColorFor returns the tcell.Color for a locked cell ID.
func ColorFor(id int) tcell.Color {
	return PieceColors[PieceID(id)]
}
