package listings

import (
 "context"
 "errors"
 "strings"
 "testing"
 "github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
 "github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
 "github.com/google/uuid"
)

func TestListingMediaReservationRecoveryIsBoundToOwnerAndBytes(t *testing.T) {
 client:=openListingClient(t);ctx:=context.Background();owner,localities,category:=createListingProvider(t,client);other,_,_:=createListingProvider(t,client)
 item,err:=NewEntRepository(client).Create(ctx,owner.ID,integrationCreate(category,localities[0]));if err!=nil{t.Fatal(err)}
 repo:=listingmedia.NewEntRepository(client)
 finder,ok:=any(repo).(interface{FindReservation(context.Context,uuid.UUID,uuid.UUID,listingmedia.UploadRequest)(uuid.UUID,string,error)})
 if !ok{t.Fatal("repository cannot recover interrupted uploads")}
 request:=listingmedia.UploadRequest{Ordinal:1,ContentType:"image/png",ByteSize:128,ChecksumSHA256:strings.Repeat("a",64)}
 id,_,err:=finder.FindReservation(ctx,owner.ID,item.ID,request);if err!=nil||id!=uuid.Nil{t.Fatal("empty slot recovery failed")}
 mediaID:=uuid.New();reference:="pending/"+mediaID.String();if err:=repo.ReservePending(ctx,owner.ID,item.ID,mediaID,request,reference);err!=nil{t.Fatal(err)}
 id,stored,err:=finder.FindReservation(ctx,owner.ID,item.ID,request);if err!=nil||id!=mediaID||stored!=reference{t.Fatalf("lost durable reservation: %v",err)}
 if _,_,err:=finder.FindReservation(ctx,other.ID,item.ID,request);!errors.Is(err,provideraccess.ErrForbidden){t.Fatal("other owner recovered capability")}
 changed:=request;changed.ChecksumSHA256=strings.Repeat("b",64)
 if _,_,err:=finder.FindReservation(ctx,owner.ID,item.ID,changed);!errors.Is(err,listingmedia.ErrConflict){t.Fatal("different bytes reused reservation")}
 if _,err:=client.Listing.UpdateOneID(item.ID).SetState("pending_review").Save(ctx);err!=nil{t.Fatal(err)}
 if _,_,err:=finder.FindReservation(ctx,owner.ID,item.ID,request);!errors.Is(err,provideraccess.ErrForbidden){t.Fatal("noneditable reservation was recovered")}
}
