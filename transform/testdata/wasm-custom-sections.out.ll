target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

!wasm.custom_sections = !{!0, !1, !2}

!0 = !{!"red", !"foo"}
!1 = !{!"green", !"bar"}
!2 = !{!"green", !"qux"}
