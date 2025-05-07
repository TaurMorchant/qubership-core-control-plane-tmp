package clustering

// ApplyOnMasterChange returns a callback that is executed only when Master node is changed
func ApplyOnMasterChange(applier func(masterNode NodeInfo, role Role)) func(NodeInfo, Role) error {
	return newOnMasterNodeChangedAggregatorCallback().AddCallback(applier).ApplyOnMasterChange
}

type onMasterNodeChangedAggregatorCallback struct {
	currentMasterNode *NodeInfo
	applyTarget       []func(NodeInfo, Role)
}

func (o *onMasterNodeChangedAggregatorCallback) ApplyOnMasterChange(newMaster NodeInfo, role Role) error {
	if o.currentMasterNode == nil { // On init
		o.currentMasterNode = &newMaster
		return nil
	}
	if o.currentMasterNode.BusAddress() == newMaster.BusAddress() &&
		o.currentMasterNode.GetHttpAddress() == newMaster.GetHttpAddress() { // Master stays the same
		return nil
	}
	o.assignThenApply(newMaster, role)
	return nil
}

func (o *onMasterNodeChangedAggregatorCallback) AddCallback(target func(NodeInfo, Role)) *onMasterNodeChangedAggregatorCallback {
	o.applyTarget = append(o.applyTarget, target)
	return o
}

func (o *onMasterNodeChangedAggregatorCallback) assignThenApply(newMaster NodeInfo, role Role) {
	o.currentMasterNode = &newMaster
	for _, functor := range o.applyTarget {
		functor(newMaster, role)
	}
}

func newOnMasterNodeChangedAggregatorCallback() *onMasterNodeChangedAggregatorCallback {
	return &onMasterNodeChangedAggregatorCallback{
		applyTarget: make([]func(NodeInfo, Role), 0),
	}
}
