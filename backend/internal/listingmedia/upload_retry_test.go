package listingmedia

import (
 "context"
 "errors"
 "testing"
 "github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
 "github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
 "github.com/google/uuid"
)

func TestUploadIntentReusesReservedObjectWithoutAnotherWrite(t *testing.T) {
 owner:=users.InternalUser{ID:uuid.New()}; mediaID:=uuid.New(); reference:="pending/"+mediaID.String()
 repo:=&retryRepository{mediaID:mediaID,reference:reference}
 storage:=&recordingStorage{reservation:StorageReservation{ObjectReference:reference,Capability:UploadCapability{URL:"https://upload.example.invalid/scoped",Method:"PUT"}}}
 intent,err:=NewService(&recordingAuthorizer{owner:owner},repo,storage).CreateUploadIntent(context.Background(),users.VerifiedIdentity{Subject:"provider"},uuid.New(),validRequest())
 if err!=nil||intent.MediaID!=mediaID||repo.calls!=0 {t.Fatalf("retry lost original reservation: id=%s writes=%d err=%v",intent.MediaID,repo.calls,err)}
}
func TestUploadIntentRejectsDifferentReservedBytesBeforeStorage(t *testing.T) {
 repo:=&retryRepository{lookupErr:ErrConflict};storage:=&recordingStorage{}
 _,err:=NewService(&recordingAuthorizer{owner:users.InternalUser{ID:uuid.New()}},repo,storage).CreateUploadIntent(context.Background(),users.VerifiedIdentity{Subject:"provider"},uuid.New(),validRequest())
 if !errors.Is(err,ErrConflict)||storage.calls!=0 {t.Fatalf("conflicting reservation reached storage: %v",err)}
 repo.lookupErr=provideraccess.ErrForbidden
 _,err=NewService(&recordingAuthorizer{owner:users.InternalUser{ID:uuid.New()}},repo,storage).CreateUploadIntent(context.Background(),users.VerifiedIdentity{Subject:"provider"},uuid.New(),validRequest())
 if !errors.Is(err,provideraccess.ErrForbidden)||storage.calls!=0 {t.Fatalf("ownership lookup was bypassed: %v",err)}
}
type retryRepository struct {recordingRepository;mediaID uuid.UUID;reference string;lookupErr error}
func(r *retryRepository) FindReservation(context.Context,uuid.UUID,uuid.UUID,UploadRequest)(uuid.UUID,string,error){return r.mediaID,r.reference,r.lookupErr}
