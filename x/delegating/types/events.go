package types

func (EventDelegate) XXX_MessageName() string { return "delegate" }

func (EventFreeze) XXX_MessageName() string { return "freeze" }

func (EventUndelegate) XXX_MessageName() string { return "undelegate" }

func (EventAccrue) XXX_MessageName() string { return "accrue" }

func (EventValidatorAccrue) XXX_MessageName() string { return "validator_accrue" }

func (EventMassiveRevoke) XXX_MessageName() string { return "massive_revoke" }
