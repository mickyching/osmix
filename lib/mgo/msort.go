package mgo

import "sort"

// Sort implement quicksort
func Sort(data sort.Interface) {
	n := data.Len()
	quickSort(data, 0, n, maxDepth(n))
}

// 2lg(n)
func maxDepth(n int) int {
	var depth int
	for i := n; i > 0; i >>= 1 {
		depth++
	}
	return depth * 2
}

func quickSort(data sort.Interface, a, b, depth int) {
	if b-a > 12 {
		depth--

		// partition too many branch, try heapsort
		if depth == 0 {
			heapSort(data, a, b)
			return
		}

		p := partition(data, a, b)
		quickSort(data, a, p, depth)
		quickSort(data, p+1, b, depth)
	}
	if b-a > 1 {
		// Do ShellSort pass with gap 6.
		// It could be written in this simplified form cause b-a <= 12.
		for i := a + 6; i < b; i++ {
			if data.Less(i, i-6) {
				data.Swap(i, i-6)
			}
		}
		insertionSort(data, a, b)
	}
}

func partition(data sort.Interface, a, b int) int {
	p, i, j := a, a, b-1
	for i < j {
		for i < j && data.Less(p, j) { // find data.j <= data.p
			j--
		}
		for i < j && !data.Less(p, i) { // find data.i > data.p
			i++
		}
		if i < j { // swap data.j <= data.p && data.i > data.p
			data.Swap(i, j)
		}
	}
	// j is partition index, so data.j should be the partion value.
	// because idx > j satisfies data.idx > data.j
	data.Swap(p, j)
	return j
}

func insertionSort(data sort.Interface, a, b int) {
	for i := a + 1; i < b; i++ {
		// sort [a, i], move data.i to right position.
		for j := i; j > a && data.Less(j, j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}

func heapSort(data sort.Interface, a, b int) {
	offset := a // offset is not changed
	b -= a
	a = 0

	// build heap with greatest element at top.
	for i := (b - 1) / 2; i >= 0; i-- {
		siftDown(data, i, b, offset)
	}

	// pop elements, largest first into end of data.
	for i := b - 1; i >= 0; i-- {
		data.Swap(offset, offset+i)
		siftDown(data, a, i, offset)
	}
}

// siftDown implements the heap property on data[a, b).
func siftDown(data sort.Interface, a, b, offset int) {
	root := a
	for {
		child := 2*root + 1

		// child out of range
		if child >= b {
			break
		}

		// set child=child+1 when data.child < data.(child+1)
		if child+1 < b && data.Less(offset+child, offset+child+1) {
			child++
		}

		// root >= child means finished
		// because sub-heap is already heapified.
		if !data.Less(offset+root, offset+child) {
			return
		}

		// root < child should sift root down.
		data.Swap(offset+root, offset+child)
		root = child
	}
}
