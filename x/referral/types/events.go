package types

func (EventStatusUpdated) XXX_MessageName() string { return "status_updated" }

func (EventStatusWillBeDowngraded) XXX_MessageName() string { return "status_will_be_downgraded" }

func (EventStatusDowngradeCanceled) XXX_MessageName() string { return "status_downgrade_canceled" }

func (EventCompression) XXX_MessageName() string { return "compression" }

func (EventTransitionRequested) XXX_MessageName() string { return "transition_requested" }

func (EventTransitionPerformed) XXX_MessageName() string { return "transition_performed" }

func (EventTransitionDeclined) XXX_MessageName() string { return "transition_declined" }

func (EventAccBanished) XXX_MessageName() string { return "acc_banished" }
