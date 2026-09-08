import { describe, expect, it } from "vitest";
import { validPhotos, validUploadIntent, validUploadRequest } from "./media-contract";

const id="aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
const photo={id,ordinal:1,width:2,height:3};
describe("listing photo contract",()=>{
  it("accepts only bounded public photo metadata",()=>{
    expect(validPhotos({photos:[photo]})).toBe(true);
    for(const value of [{photos:[{...photo,objectReference:"private/key"}]},{photos:[photo,photo]},{photos:[{...photo,width:8193}]},{photos:[{...photo,width:8192,height:8192}]}]) expect(validPhotos(value)).toBe(false);
  });
  it("accepts only a scoped upload capability from configured storage",()=>{
    const value={mediaId:id,capability:{url:`https://storage.example.test/storage/v1/object/upload/sign/listing-images/pending/${id}?token=synthetic-upload-token`,method:"PUT",headers:{"Content-Type":"image/png"}}};
    expect(validUploadIntent(value,"https://storage.example.test")).toBe(true);
    expect(validUploadIntent(value,"https://other.example.test")).toBe(false);
    expect(validUploadIntent({...value,capability:{...value.capability,headers:{Authorization:"secret"}}},"https://storage.example.test")).toBe(false);
    expect(validUploadIntent({...value,capability:{...value.capability,url:value.capability.url.replace(id,"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")}},"https://storage.example.test")).toBe(false);
  });
  it("does not accept client publication metadata or oversized uploads",()=>{
    const value={ordinal:1,contentType:"image/png",byteSize:42,checksumSha256:"a".repeat(64)};
    expect(validUploadRequest(value)).toBe(true);
    expect(validUploadRequest({...value,state:"ready"})).toBe(false);
    expect(validUploadRequest({...value,byteSize:10485761})).toBe(false);
  });
});
