package types

// Generates reports whether the element 'from' generates the element 'to'
// in the productive cycle (相生: 木→火→土→金→水→木).
func Generates(from, to Element) bool {
	return generatesTarget[from] == to
}

// Overcomes reports whether the element 'from' overcomes the element 'to'
// in the destructive cycle (相克: 木→土→水→火→金→木).
func Overcomes(from, to Element) bool {
	return overcomesTarget[from] == to
}

// generatesTarget maps each element to the element it generates.
var generatesTarget = [ElementCount]Element{
	Wood:  Fire,
	Fire:  Earth,
	Earth: Metal,
	Metal: Water,
	Water: Wood,
}

// overcomesTarget maps each element to the element it overcomes.
var overcomesTarget = [ElementCount]Element{
	Wood:  Earth,
	Fire:  Metal,
	Earth: Water,
	Metal: Wood,
	Water: Fire,
}
