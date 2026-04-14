(globalThis.TURBOPACK||(globalThis.TURBOPACK=[])).push(["object"==typeof document?document.currentScript:void 0,86238,e=>{"use strict";var o=e.i(43476),t=e.i(10007),i=e.i(23177),n=e.i(51892),r=e.i(22305),l=e.i(27828),s=e.i(36730),d=e.i(32322),a=e.i(71645);let c=d.default.icon({iconUrl:"https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-gold.png",iconRetinaUrl:"https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-gold.png",shadowUrl:"https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png",iconSize:[25,41],iconAnchor:[12,41],popupAnchor:[1,-34],shadowSize:[41,41]});function p({center:e}){let o=(0,s.useMap)(),[t,i]=(0,a.useState)(!0);return(0,a.useEffect)(()=>{let e=()=>i(!1);return o.on("dragstart",e),()=>{o.off("dragstart",e)}},[o]),(0,a.useEffect)(()=>{t&&0!==e[0]&&0!==e[1]&&o.setView(e,o.getZoom(),{animate:!0,duration:.5})},[e,o,t]),null}function x({color:e,border:t,label:i,pulse:n}){return(0,o.jsxs)("div",{style:{display:"flex",alignItems:"center",gap:"6px"},children:[(0,o.jsx)("div",{style:{width:"11px",height:"11px",borderRadius:"50%",background:e,border:t?`2px solid ${t}`:"none",boxShadow:n?"0 0 6px #00e5cc":"none",flexShrink:0}}),(0,o.jsx)("span",{style:{fontSize:"9px",color:"var(--ink-secondary)",letterSpacing:"0.5px"},children:i})]})}e.s(["default",0,function({dronePos:e,waypoints:s,startPoint:h,currentWaypointIndex:g,replayTrajectory:f}){let b=[e.lat,e.lng],u=(0,a.useMemo)(()=>{let e=[];if(s.length<2)return e;for(let o=0;o<s.length-1;o++){let t=s[o],i=s[o+1],n=i.index<=g;e.push({positions:[[t.lat,t.lng],[i.lat,i.lng]],color:n?"#22c55e":"#00e5cc",dashed:!n})}return e},[s,g]),m=s.filter(e=>e.index<g).length;return(0,o.jsxs)("div",{style:{width:"100%",height:"100%",position:"relative",borderRadius:"8px",overflow:"hidden",border:"2px solid var(--hud-cyan)"},children:[(0,o.jsxs)(t.MapContainer,{center:0!==b[0]?b:[10.4806,-66.9036],zoom:18,style:{height:"100%",width:"100%"},zoomControl:!1,children:[(0,o.jsx)(i.TileLayer,{url:"https://mt1.google.com/vt/lyrs=y&x={x}&y={y}&z={z}",attribution:"© Google Maps",maxZoom:22}),(0,o.jsx)(p,{center:b}),f&&f.length>1&&(0,o.jsx)(l.Polyline,{positions:f,color:"#fbbf24",weight:3,opacity:.75,dashArray:void 0}),u.map((e,t)=>(0,o.jsx)(l.Polyline,{positions:e.positions,color:e.color,weight:3,dashArray:e.dashed?"8, 8":void 0,opacity:.85},t)),h&&(0,o.jsx)(n.Marker,{position:[h.lat,h.lng],icon:d.default.divIcon({className:"",html:`
      <div style="
        width:30px;
        height:30px;
        border-radius:50%;
        background:rgba(2,4,10,0.85);
        border:2.5px solid #00e5cc;
        color:#00e5cc;
        font-size:13px;
        font-weight:700;
        font-family:monospace;
        display:flex;
        align-items:center;
        justify-content:center;
        box-shadow:none;
      ">S</div>
    `,iconSize:[30,30],iconAnchor:[15,15],popupAnchor:[0,-19]}),zIndexOffset:500,children:(0,o.jsx)(r.Popup,{children:(0,o.jsxs)("div",{style:{fontFamily:"var(--font-mono)",fontSize:"11px",background:"#02040a",color:"#fff",padding:"6px",borderRadius:"4px"},children:[(0,o.jsx)("strong",{style:{color:"#22c55e"},children:"🏠 HOME / TAKEOFF"}),(0,o.jsx)("br",{}),"LAT: ",h.lat.toFixed(6),(0,o.jsx)("br",{}),"LNG: ",h.lng.toFixed(6),(0,o.jsx)("br",{}),"ALT: ",h.alt,"m"]})})}),s.map(e=>{var t,i;let l,s,a=(t=e.index)<g?"completed":t===g?"active":"pending",c=(i=e.index,l=({completed:{bg:"#22c55e",border:"#16a34a",text:"#fff"},active:{bg:"#00e5cc",border:"#00b5a0",text:"#000"},pending:{bg:"rgba(2,4,10,0.85)",border:"#00e5cc",text:"#00e5cc"}})[a],s="active"===a?30:26,d.default.divIcon({className:"",html:`
      <div style="
        width:${s}px;
        height:${s}px;
        border-radius:50%;
        background:${l.bg};
        border:2.5px solid ${l.border};
        color:${l.text};
        font-size:${i>=100?8:i>=10?10:12}px;
        font-weight:700;
        font-family:monospace;
        display:flex;
        align-items:center;
        justify-content:center;
        box-shadow:${"active"===a?`0 0 10px ${l.border}, 0 0 20px ${l.border}`:"completed"===a?"0 0 6px #22c55e":"none"};
        transition:all 0.3s ease;
        pointer-events:auto;
      ">${i}</div>
    `,iconSize:[s,s],iconAnchor:[s/2,s/2],popupAnchor:[0,-(s/2+4)]}));return(0,o.jsx)(n.Marker,{position:[e.lat,e.lng],icon:c,children:(0,o.jsx)(r.Popup,{children:(0,o.jsxs)("div",{style:{fontFamily:"var(--font-mono)",fontSize:"11px",background:"#02040a",color:"#fff",padding:"6px",borderRadius:"4px"},children:[(0,o.jsxs)("strong",{style:{color:"completed"===a?"#22c55e":"active"===a?"#00e5cc":"#fff"},children:["WP ",e.index]}),(0,o.jsx)("br",{}),"ALT: ",e.alt,"m",(0,o.jsx)("br",{}),"completed"===a?"✅ COMPLETADO":"active"===a?"📍 EN RUTA":"⏳ PENDIENTE"]})})},e.index)}),0!==b[0]&&c&&(0,o.jsx)(n.Marker,{position:b,icon:c,zIndexOffset:1e3,children:(0,o.jsx)(r.Popup,{children:(0,o.jsxs)("div",{style:{fontFamily:"var(--font-mono)",fontSize:"11px",background:"#02040a",color:"#fff",padding:"6px",borderRadius:"4px"},children:[(0,o.jsx)("strong",{style:{color:"#fbbf24"},children:"DRON POSICIÓN"}),(0,o.jsx)("br",{}),"LAT: ",e.lat.toFixed(6),(0,o.jsx)("br",{}),"LNG: ",e.lng.toFixed(6)]})})})]}),(0,o.jsxs)("div",{style:{position:"absolute",bottom:10,right:10,background:"rgba(2, 4, 10, 0.85)",padding:"8px 12px",borderRadius:"6px",border:"1px solid var(--hud-cyan)",zIndex:1e3,pointerEvents:"none",display:"flex",flexDirection:"column",gap:"2px"},children:[(0,o.jsx)("div",{style:{fontSize:"9px",color:"var(--ink-tertiary)",letterSpacing:"1px"},children:"MODO MAPA TÁCTICO"}),(0,o.jsx)("div",{style:{fontSize:"11px",color:"var(--hud-cyan)",fontWeight:"bold"},children:"SIGUIENDO UNIDAD"}),s.length>0&&(0,o.jsxs)("div",{style:{fontSize:"10px",color:"#22c55e",fontWeight:"bold",marginTop:"2px"},children:["WP ",g,"/",s.length," •"," ",(0,o.jsxs)("span",{style:{color:"#22c55e"},children:[m," OK"]})]})]}),s.length>0&&(0,o.jsxs)("div",{style:{position:"absolute",bottom:10,left:10,background:"rgba(2, 4, 10, 0.85)",padding:"8px 10px",borderRadius:"6px",border:"1px solid rgba(255,255,255,0.1)",zIndex:1e3,pointerEvents:"none",display:"flex",flexDirection:"column",gap:"5px"},children:[(0,o.jsx)(x,{color:"#22c55e",label:"Completado"}),(0,o.jsx)(x,{color:"#00e5cc",label:"En ruta",pulse:!0}),(0,o.jsx)(x,{color:"rgba(2,4,10,0.85)",border:"#00e5cc",label:"Pendiente"})]})]})}])},24961,e=>{e.n(e.i(86238))}]);