package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"user-service/internal/domain"
	"user-service/internal/repository"
)

// WorkspaceService handles workspace business logic
type WorkspaceService struct {
	workspaceRepo *repository.WorkspaceRepository
	memberRepo    *repository.WorkspaceMemberRepository
	joinReqRepo   *repository.JoinRequestRepository
	profileRepo   *repository.UserProfileRepository
	userRepo      *repository.UserRepository
	logger        *zap.Logger
}

// NewWorkspaceService creates a new WorkspaceService
func NewWorkspaceService(
	workspaceRepo *repository.WorkspaceRepository,
	memberRepo *repository.WorkspaceMemberRepository,
	joinReqRepo *repository.JoinRequestRepository,
	profileRepo *repository.UserProfileRepository,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
		joinReqRepo:   joinReqRepo,
		profileRepo:   profileRepo,
		userRepo:      userRepo,
		logger:        logger,
	}
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(ownerID uuid.UUID, req domain.CreateWorkspaceRequest) (*domain.Workspace, error) {
	// 💡 Check if user exists before creating workspace
	// This prevents FK constraint violation if user sync failed during OAuth login
	exists, err := s.userRepo.Exists(ownerID)
	if err != nil {
		s.logger.Error("Failed to check user existence", zap.Error(err))
		return nil, errors.New("failed to verify user, please try logging in again")
	}
	if !exists {
		s.logger.Warn("User not found in database, OAuth sync may have failed", zap.String("userId", ownerID.String()))
		return nil, errors.New("user not found, please log out and log in again to sync your account")
	}

	// Default values: all true
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	needApproved := true
	if req.NeedApproved != nil {
		needApproved = *req.NeedApproved
	}

	workspace := &domain.Workspace{
		ID:                   uuid.New(),
		OwnerID:              ownerID,
		WorkspaceName:        req.WorkspaceName,
		WorkspaceDescription: req.WorkspaceDescription,
		IsPublic:             isPublic,
		NeedApproved:         needApproved,
		OnlyOwnerCanInvite:   true,
		IsActive:             true,
		CreatedAt:            time.Now(),
	}

	if err := s.workspaceRepo.Create(workspace); err != nil {
		s.logger.Error("Failed to create workspace", zap.Error(err))
		return nil, err
	}

	// Add owner as a member
	member := &domain.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
		RoleName:    domain.RoleOwner,
		IsDefault:   true,
		IsActive:    true,
		JoinedAt:    time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.memberRepo.Create(member); err != nil {
		s.logger.Error("Failed to create workspace member", zap.Error(err))
		// Should we rollback workspace creation?
	}

	// Create profile for owner
	user, err := s.userRepo.FindByID(ownerID)
	if err == nil {
		// Use user's name as default nickname, fallback to email if name is empty
		defaultNickName := user.Name
		if defaultNickName == "" {
			defaultNickName = user.Email
		}
		profile := &domain.UserProfile{
			ID:          uuid.New(),
			UserID:      ownerID,
			WorkspaceID: workspace.ID,
			NickName:    defaultNickName,
			Email:       user.Email,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.profileRepo.Create(profile); err != nil {
			s.logger.Error("Failed to create profile for owner", zap.Error(err))
		}
	}

	s.logger.Info("Workspace created", zap.String("workspaceId", workspace.ID.String()))
	return workspace, nil
}

// GetWorkspace gets a workspace by ID
func (s *WorkspaceService) GetWorkspace(id uuid.UUID) (*domain.Workspace, error) {
	return s.workspaceRepo.FindByID(id)
}

// GetWorkspaceWithOwner gets a workspace with owner
func (s *WorkspaceService) GetWorkspaceWithOwner(id uuid.UUID) (*domain.Workspace, error) {
	return s.workspaceRepo.FindByIDWithOwner(id)
}

// GetUserWorkspaces gets all workspaces for a user
func (s *WorkspaceService) GetUserWorkspaces(userID uuid.UUID) ([]domain.WorkspaceMember, error) {
	return s.memberRepo.FindByUser(userID)
}

// GetWorkspacesByOwner gets workspaces by owner
func (s *WorkspaceService) GetWorkspacesByOwner(ownerID uuid.UUID) ([]domain.Workspace, error) {
	return s.workspaceRepo.FindByOwnerID(ownerID)
}

// SearchPublicWorkspaces searches public workspaces by name
func (s *WorkspaceService) SearchPublicWorkspaces(name string) ([]domain.Workspace, error) {
	return s.workspaceRepo.FindPublicByName(name)
}

// UpdateWorkspace updates a workspace
func (s *WorkspaceService) UpdateWorkspace(id uuid.UUID, userID uuid.UUID, req domain.UpdateWorkspaceRequest) (*domain.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if user is owner
	if workspace.OwnerID != userID {
		return nil, errors.New("only owner can update workspace")
	}

	if req.WorkspaceName != nil {
		workspace.WorkspaceName = *req.WorkspaceName
	}
	if req.WorkspaceDescription != nil {
		workspace.WorkspaceDescription = req.WorkspaceDescription
	}
	if req.IsPublic != nil {
		workspace.IsPublic = *req.IsPublic
	}
	if req.NeedApproved != nil {
		workspace.NeedApproved = *req.NeedApproved
	}

	if err := s.workspaceRepo.Update(workspace); err != nil {
		s.logger.Error("Failed to update workspace", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Workspace updated", zap.String("workspaceId", workspace.ID.String()))
	return workspace, nil
}

// UpdateWorkspaceSettings updates workspace settings
func (s *WorkspaceService) UpdateWorkspaceSettings(id uuid.UUID, userID uuid.UUID, req domain.UpdateWorkspaceSettingsRequest) (*domain.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if user is owner
	if workspace.OwnerID != userID {
		return nil, errors.New("only owner can update workspace settings")
	}

	if req.WorkspaceName != nil {
		workspace.WorkspaceName = *req.WorkspaceName
	}
	if req.WorkspaceDescription != nil {
		workspace.WorkspaceDescription = req.WorkspaceDescription
	}
	if req.IsPublic != nil {
		workspace.IsPublic = *req.IsPublic
	}
	if req.RequiresApproval != nil {
		workspace.NeedApproved = *req.RequiresApproval
	}
	if req.OnlyOwnerCanInvite != nil {
		workspace.OnlyOwnerCanInvite = *req.OnlyOwnerCanInvite
	}

	if err := s.workspaceRepo.Update(workspace); err != nil {
		s.logger.Error("Failed to update workspace settings", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Workspace settings updated", zap.String("workspaceId", workspace.ID.String()))
	return workspace, nil
}

// DeleteWorkspace soft deletes a workspace
func (s *WorkspaceService) DeleteWorkspace(id uuid.UUID, userID uuid.UUID) error {
	workspace, err := s.workspaceRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Check if user is owner
	if workspace.OwnerID != userID {
		return errors.New("only owner can delete workspace")
	}

	if err := s.workspaceRepo.SoftDelete(id); err != nil {
		s.logger.Error("Failed to delete workspace", zap.Error(err))
		return err
	}

	s.logger.Info("Workspace deleted", zap.String("workspaceId", id.String()))
	return nil
}

// SetDefaultWorkspace sets a workspace as default for user
func (s *WorkspaceService) SetDefaultWorkspace(userID, workspaceID uuid.UUID) error {
	// Check if user is a member
	isMember, err := s.memberRepo.IsMember(workspaceID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this workspace")
	}

	return s.memberRepo.SetDefault(userID, workspaceID)
}

// GetMembers gets all members of a workspace
func (s *WorkspaceService) GetMembers(workspaceID uuid.UUID) ([]domain.WorkspaceMember, error) {
	return s.memberRepo.FindByWorkspace(workspaceID)
}

// GetMembersWithProfiles gets all members of a workspace with profile info
func (s *WorkspaceService) GetMembersWithProfiles(workspaceID uuid.UUID) ([]domain.WorkspaceMemberResponse, error) {
	members, err := s.memberRepo.FindByWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	// Get all profiles for this workspace
	profiles, err := s.profileRepo.FindByWorkspace(workspaceID)
	if err != nil {
		s.logger.Warn("Failed to fetch profiles for workspace", zap.Error(err))
		// Continue without profiles
	}

	// Create a map of userID -> profile
	profileMap := make(map[uuid.UUID]*domain.UserProfile)
	for i := range profiles {
		profileMap[profiles[i].UserID] = &profiles[i]
	}

	// Build responses with profile info
	responses := make([]domain.WorkspaceMemberResponse, len(members))
	for i, m := range members {
		resp := m.ToResponse()

		// Add profile info if available
		if profile, ok := profileMap[m.UserID]; ok {
			resp.NickName = profile.NickName
			resp.UserEmail = profile.Email
			if profile.ProfileImageURL != nil {
				resp.ProfileImageUrl = *profile.ProfileImageURL
			}
		}

		responses[i] = resp
	}

	return responses, nil
}

// InviteMember invites a user to a workspace
func (s *WorkspaceService) InviteMember(workspaceID, inviterID uuid.UUID, req domain.InviteMemberRequest) (*domain.WorkspaceMember, error) {
	// Check if inviter has permission (OWNER or ADMIN)
	role, err := s.memberRepo.GetRole(workspaceID, inviterID)
	if err != nil {
		return nil, errors.New("inviter is not a member of this workspace")
	}
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return nil, errors.New("only owner or admin can invite members")
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if already a member
	isMember, err := s.memberRepo.IsMember(workspaceID, user.ID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("user is already a member of this workspace")
	}

	// Create member
	roleName := domain.RoleMember
	if req.RoleName != "" {
		roleName = req.RoleName
	}

	member := &domain.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		RoleName:    roleName,
		IsActive:    true,
		JoinedAt:    time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.memberRepo.Create(member); err != nil {
		s.logger.Error("Failed to create member", zap.Error(err))
		return nil, err
	}

	// Create profile for new member
	// Use user's name as default nickname, fallback to email if name is empty
	memberNickName := user.Name
	if memberNickName == "" {
		memberNickName = user.Email
	}
	profile := &domain.UserProfile{
		ID:          uuid.New(),
		UserID:      user.ID,
		WorkspaceID: workspaceID,
		NickName:    memberNickName,
		Email:       user.Email,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.profileRepo.Create(profile); err != nil {
		s.logger.Error("Failed to create profile for member", zap.Error(err))
	}

	member.User = user
	s.logger.Info("Member invited", zap.String("workspaceId", workspaceID.String()), zap.String("userId", user.ID.String()))
	return member, nil
}

// UpdateMemberRole updates a member's role
func (s *WorkspaceService) UpdateMemberRole(workspaceID, memberID, updaterID uuid.UUID, req domain.UpdateMemberRoleRequest) (*domain.WorkspaceMember, error) {
	// Check if updater is owner
	role, err := s.memberRepo.GetRole(workspaceID, updaterID)
	if err != nil {
		return nil, errors.New("updater is not a member of this workspace")
	}
	if role != domain.RoleOwner {
		return nil, errors.New("only owner can update member roles")
	}

	member, err := s.memberRepo.FindByID(memberID)
	if err != nil {
		return nil, err
	}

	if member.WorkspaceID != workspaceID {
		return nil, errors.New("member not found in this workspace")
	}

	// Cannot change owner's role
	if member.RoleName == domain.RoleOwner {
		return nil, errors.New("cannot change owner's role")
	}

	member.RoleName = req.RoleName
	member.UpdatedAt = time.Now()

	if err := s.memberRepo.Update(member); err != nil {
		s.logger.Error("Failed to update member role", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Member role updated", zap.String("memberId", memberID.String()))
	return member, nil
}

// RemoveMember removes a member from a workspace
func (s *WorkspaceService) RemoveMember(workspaceID, memberID, removerID uuid.UUID) error {
	// Check if remover has permission
	role, err := s.memberRepo.GetRole(workspaceID, removerID)
	if err != nil {
		return errors.New("remover is not a member of this workspace")
	}
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return errors.New("only owner or admin can remove members")
	}

	member, err := s.memberRepo.FindByID(memberID)
	if err != nil {
		return err
	}

	if member.WorkspaceID != workspaceID {
		return errors.New("member not found in this workspace")
	}

	// Cannot remove owner
	if member.RoleName == domain.RoleOwner {
		return errors.New("cannot remove owner")
	}

	// Admin can only remove regular members
	if role == domain.RoleAdmin && member.RoleName == domain.RoleAdmin {
		return errors.New("admin cannot remove other admins")
	}

	if err := s.memberRepo.Delete(memberID); err != nil {
		s.logger.Error("Failed to remove member", zap.Error(err))
		return err
	}

	s.logger.Info("Member removed", zap.String("memberId", memberID.String()))
	return nil
}

// IsMember checks if user is a member of workspace
func (s *WorkspaceService) IsMember(workspaceID, userID uuid.UUID) (bool, error) {
	return s.memberRepo.IsMember(workspaceID, userID)
}

// ValidateMemberAccess validates if user has access to workspace
func (s *WorkspaceService) ValidateMemberAccess(workspaceID, userID uuid.UUID) (bool, error) {
	return s.memberRepo.IsMember(workspaceID, userID)
}

// CreateJoinRequest creates a join request
func (s *WorkspaceService) CreateJoinRequest(workspaceID, userID uuid.UUID) (*domain.WorkspaceJoinRequest, error) {
	// Check if workspace exists and is public
	workspace, err := s.workspaceRepo.FindByID(workspaceID)
	if err != nil {
		return nil, err
	}
	if !workspace.IsPublic {
		return nil, errors.New("workspace is not public")
	}

	// Check if already a member
	isMember, err := s.memberRepo.IsMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("already a member of this workspace")
	}

	// Check if already has pending request
	hasPending, err := s.joinReqRepo.HasPendingRequest(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, errors.New("already has a pending join request")
	}

	request := &domain.WorkspaceJoinRequest{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Status:      domain.JoinStatusPending,
		RequestedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.joinReqRepo.Create(request); err != nil {
		s.logger.Error("Failed to create join request", zap.Error(err))
		return nil, err
	}

	// If no approval needed, approve immediately
	if !workspace.NeedApproved {
		return s.ProcessJoinRequest(workspaceID, request.ID, workspace.OwnerID, domain.ProcessJoinRequestRequest{
			Status: domain.JoinStatusApproved,
		})
	}

	s.logger.Info("Join request created", zap.String("requestId", request.ID.String()))
	return request, nil
}

// GetJoinRequests gets join requests for a workspace
func (s *WorkspaceService) GetJoinRequests(workspaceID, userID uuid.UUID) ([]domain.WorkspaceJoinRequest, error) {
	// Check if user has permission
	role, err := s.memberRepo.GetRole(workspaceID, userID)
	if err != nil {
		return nil, errors.New("user is not a member of this workspace")
	}
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return nil, errors.New("only owner or admin can view join requests")
	}

	return s.joinReqRepo.FindByWorkspace(workspaceID)
}

// GetPendingJoinRequests gets pending join requests for a workspace
func (s *WorkspaceService) GetPendingJoinRequests(workspaceID, userID uuid.UUID) ([]domain.WorkspaceJoinRequest, error) {
	// Check if user has permission
	role, err := s.memberRepo.GetRole(workspaceID, userID)
	if err != nil {
		return nil, errors.New("user is not a member of this workspace")
	}
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return nil, errors.New("only owner or admin can view join requests")
	}

	return s.joinReqRepo.FindPendingByWorkspace(workspaceID)
}

// ProcessJoinRequest processes a join request (approve/reject)
func (s *WorkspaceService) ProcessJoinRequest(workspaceID, requestID, processorID uuid.UUID, req domain.ProcessJoinRequestRequest) (*domain.WorkspaceJoinRequest, error) {
	// Check if processor has permission
	role, err := s.memberRepo.GetRole(workspaceID, processorID)
	if err != nil {
		return nil, errors.New("processor is not a member of this workspace")
	}
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return nil, errors.New("only owner or admin can process join requests")
	}

	request, err := s.joinReqRepo.FindByID(requestID)
	if err != nil {
		return nil, err
	}

	if request.WorkspaceID != workspaceID {
		return nil, errors.New("join request not found for this workspace")
	}

	if request.Status != domain.JoinStatusPending {
		return nil, errors.New("join request is not pending")
	}

	request.Status = req.Status
	request.UpdatedAt = time.Now()

	if err := s.joinReqRepo.Update(request); err != nil {
		s.logger.Error("Failed to update join request", zap.Error(err))
		return nil, err
	}

	// If approved, add as member
	if req.Status == domain.JoinStatusApproved {
		member := &domain.WorkspaceMember{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			UserID:      request.UserID,
			RoleName:    domain.RoleMember,
			IsActive:    true,
			JoinedAt:    time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.memberRepo.Create(member); err != nil {
			s.logger.Error("Failed to create member from join request", zap.Error(err))
			return nil, err
		}

		// Create profile for new member
		user, err := s.userRepo.FindByID(request.UserID)
		if err == nil {
			// Use user's name as default nickname, fallback to email if name is empty
			approvedNickName := user.Name
			if approvedNickName == "" {
				approvedNickName = user.Email
			}
			profile := &domain.UserProfile{
				ID:          uuid.New(),
				UserID:      request.UserID,
				WorkspaceID: workspaceID,
				NickName:    approvedNickName,
				Email:       user.Email,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := s.profileRepo.Create(profile); err != nil {
				s.logger.Error("Failed to create profile for member", zap.Error(err))
			}
		}
	}

	s.logger.Info("Join request processed", zap.String("requestId", requestID.String()), zap.String("status", string(req.Status)))
	return request, nil
}
