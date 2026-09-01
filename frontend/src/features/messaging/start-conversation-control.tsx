"use client";

import { useState } from "react";

export function StartConversationControl({listingId,label,error,messagesUrl}:{listingId:string;label:string;error:string;messagesUrl:string}){
 const [pending,setPending]=useState(false),[failed,setFailed]=useState(false);
 async function start(){if(pending)return;setPending(true);setFailed(false);try{const response=await fetch("/api/v1/me/conversations",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({listingId})});const value:unknown=await response.json();if(!response.ok||!validConversation(value))throw new Error();window.location.assign(`${messagesUrl}?conversation=${value.id}`)}catch{setFailed(true)}finally{setPending(false)}}
 return <div className="mt-3"><button type="button" className="market-button-secondary w-full" disabled={pending} onClick={()=>void start()}>{label}</button>{failed?<p role="alert" className="mt-3 text-earth">{error}</p>:null}</div>
}
function validConversation(value:unknown):value is {id:string}{return value!==null&&typeof value==="object"&&!Array.isArray(value)&&typeof (value as Record<string,unknown>).id==="string"&&/^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test((value as {id:string}).id)}
