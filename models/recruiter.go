// If you need to add these fields to your existing Recruiter struct, add the following fields:

// Status of the recruiter (pending, approved, rejected)
Status string `json:"status" bson:"status" default:"pending"`

// AdminVerified indicates if the recruiter is verified by an admin
AdminVerified bool `json:"adminVerified" bson:"adminVerified" default:"false"`

// RejectionReason provides a reason if the recruiter was rejected
RejectionReason string `json:"rejectionReason,omitempty" bson:"rejectionReason,omitempty"`

// ApprovedBy stores the admin ID who approved/rejected the recruiter
ApprovedBy primitive.ObjectID `json:"approvedBy,omitempty" bson:"approvedBy,omitempty"`

// ApprovalDate stores when the recruiter was approved/rejected
ApprovalDate time.Time `json:"approvalDate,omitempty" bson:"approvalDate,omitempty"`
