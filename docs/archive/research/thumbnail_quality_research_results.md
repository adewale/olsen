# Thumbnail Quality Optimization – v2 Technical Spec (with Instrumentation & Correctness Guardrails)
_Last updated: 2025-10-11 15:14 UTC_

## 0) Purpose of this Revision
This v2 adds a **full instrumentation and correctness strategy** to the original report so we can **prove** whether quality is being degraded by bugs, bad defaults, or missing steps. Fixing broken/missing functionality often delivers the **largest wins**, so the plan below makes those problems obvious and quantifiable.

---

## 1) Executive Summary (Blunt)
- **Instrument the pipeline end-to-end** (decode → orient → colorspace → resize → sharpen → encode → store) and **surface hard evidence** when anything silently degrades quality.
- Add **stage timing, counters, and structured logs**; capture **intermediate artifacts** (sampled) for human inspection; ship **SSIM/PSNR/Laplacian** checks into CI and prod sampling.
- Add **guardrails**: assert invariants (no upscaling unless explicit; EXIF orientation applied exactly once; ICC handled; chroma subsampling as intended; no half-size RAW decode unless opted-in; gamma path consistent).
- Track **fallback rates** (embedded JPEG, non-CGO, decode failures) and **quality deltas** when fallbacks happen. If we’re doing something wrong, we’ll see it, with repros.
- Priority: **correctness before cleverness**. Fix the bugs and omissions, then tune filters/encoders.

---

## 2) “Quality Gotchas” We Must Catch
These are the common, high-impact mistakes that silently trash image quality. The instrumentation must make each one visible:
1. **Wrong source chosen**: Using an embedded JPEG instead of full RAW without logging; or upscaling a tiny embedded preview to 1024px.
2. **EXIF orientation mishandled**: Not applied, applied twice, or applied after compression (extra resample).
3. **Color management missing/misapplied**: Treating AdobeRGB/Display P3 as sRGB; dropping ICC; double or missing gamma conversion.
4. **Premature precision loss**: For RAW, decoding to 8-bit and resizing in 8-bit when 16-bit→resize→8-bit is feasible.
5. **Unintended half-res paths**: LibRaw `half_size` or similar “fast path” used unintentionally.
6. **Resizing pitfalls**: Accidental nearest-neighbor; double-resize (resize an already resized image); resize before orientation/color; failure to do resizing in linear light (if we choose that policy).
7. **Encoder mistakes**: Defaulting to low JPEG quality; unknown/suboptimal chroma subsampling (e.g., 4:2:0 blurring fine edges/text where 4:4:4 is expected); unexpected progressive/baseline flips.
8. **Concurrency/race issues**: Buffer reuse causing corruption; data races that intermittently produce softer/odd thumbnails.
9. **Alpha/gray/indexed oddities**: Dropping alpha incorrectly, wrong premultiplication; grayscale misinterpreted as YCbCr; palette images decoded poorly.
10. **Storage truncation**: BLOB truncated or re-encoded elsewhere; unexpected recompression during transport or caching layers.

---

## 3) What to Instrument (Signals & Artifacts)

### 3.1 Per-Image Diagnostics (Structured Log, JSON)
Emit **one JSON object per image** (attach to job context) with fields below. Store in rolling log + optional sidecar JSON when sampling is on.

```json
{
  "img_id": "sha256:... or path",
  "source": {
    "format": "RAW|JPEG|PNG|TIFF|...",
    "input_w": 6048, "input_h": 4024,
    "has_icc": true, "icc_desc": "AdobeRGB", "exif_orientation": 6,
    "raw": {
      "libraw_enabled": true,
      "demosaic": "AHD|DCB|PPG|unknown",
      "output_bps": 16,
      "output_color": "sRGB|AdobeRGB|linear|unknown",
      "use_camera_wb": true,
      "half_size": false
    },
    "fallback_reason": "none|no_cgo|decode_error|no_raw|embedded_only"
  },
  "pipeline": {
    "orientation_applied": true,
    "colorspace_in": "sRGB|AdobeRGB|linear|unknown",
    "colorspace_out": "sRGB",
    "gamma_linearized": true,
    "resize": {
      "target_long_edge": 1024,
      "filter": "Lanczos3|Lanczos2|Mitchell|CatmullRom|Bilinear",
      "pre_sharpen": {"enabled": true, "amount": 0.5, "radius": 0.8},
      "post_sharpen": {"enabled": true, "amount": 0.3, "radius": 0.5},
      "upscale": false
    },
    "encode": {
      "format": "jpeg|webp|avif",
      "quality": 85,
      "chroma": "444|420|422",
      "progressive": false,
      "bytes": 153240
    }
  },
  "metrics": {
    "ssim_vs_ref": 0.962,
    "psnr_vs_ref_db": 36.8,
    "lap_var": 215.3,
    "delta_e_mean": 1.8,
    "clipped_pixels": {"low": 112, "high": 94},
    "histogram_banding_score": 0.01
  },
  "timing_ms": {
    "decode": 41.2, "orient": 0.4, "color": 1.8, "resize": 3.6, "sharpen": 1.5, "encode": 5.1, "store": 0.7, "total": 54.3
  },
  "warnings": ["upscale_detected", "icc_missing_assumed_srgb", "half_size_decode_suspected"],
  "version": {
    "thumb_pipeline": "2.0.0",
    "libraw": "0.21.x",
    "encoder": "stdjpeg|libjpeg-turbo|libwebp|libavif"
  }
}
```

**Purpose:** Makes silent failures obvious. If `fallback_reason != none`, we know quality constraints up front. If `upscale=true`, that’s a bug or at least a quality red flag. If `icc_missing`, we explicitly record the assumption.

### 3.2 Prometheus/expvar Counters & Histograms
Expose metrics for dashboards and alerting:
- **Rates/Counters**
  - `thumb_fallback_total{reason}` (no_cgo, decode_error, embedded_only, etc.)
  - `thumb_orientation_double_apply_total` (should be 0)
  - `thumb_upscale_total` (should be 0 in default policy)
  - `thumb_colorspace_mismatch_total` (non-sRGB in → sRGB out without proper transform)
  - `thumb_chroma_420_total`, `thumb_chroma_444_total`
  - `thumb_quality_bucket_total{q=75|80|85|90|95}`
- **Histograms/Summaries**
  - `thumb_stage_ms{stage=decode|orient|resize|encode|total}`
  - `thumb_bytes{size=64|256|512|1024}`
  - `thumb_ssim_vs_ref` (sampled)
  - `thumb_delta_e_mean` (sampled)

Alert on spikes, e.g., `thumb_fallback_total{reason="no_cgo"}` rising, or `thumb_upscale_total > 0`.

### 3.3 Intermediate Artifact Capture (Sampled)
When `THUMB_QA_SAMPLE=1/100` (1%), **persist intermediates** for repro/debug:
- `*_decode.png` (post-decode, pre-orientation/color)
- `*_after_orient_color.png`
- `*_resized.png`
- `*_final.jpg|webp|avif`
- `*_diag.json` (the JSON above)

Retention is bounded (e.g., cap N artifacts or days). Store under `~/.olsen/qa/{date}/...` or temp dir.

### 3.4 Reference Generation (for Metrics)
When sampling is enabled, generate a **reference thumbnail** per target size using a high-fidelity path (e.g., linear-light Lanczos3 or a fixed “gold” path) and compute **SSIM/PSNR/ΔE** against it. Keep only the scores in-prod; retain images only under sampling.

---

## 4) Guardrails (Hard Checks & Assertions)

- **No unintended upscales**: If `input_long_edge < target_long_edge`, either **refuse** or downrank quality expectations; emit warning.
- **EXIF orientation applied exactly once**: Track orientation state; assert no double-rotate. Unit tests cover all 8 orientations.
- **ICC/Color**: If ICC present and not sRGB, **convert** to sRGB before resizing/encoding; if ICC missing, assume sRGB and log `icc_missing_assumed_srgb` once per image.
- **Gamma policy**: If linear-light resampling enabled, enforce `{linearize -> resize -> delinearize}`; assert the flags reflect that path.
- **RAW decode policy**: Assert `half_size=false` unless explicitly configured. Record demosaic, WB, output_bps; warn if 8-bit path used where 16-bit is supported.
- **Encoder policy**: Log chroma subsampling. If 4:4:4 is required for text/graphics thumbs, assert policy per MIME (e.g., screenshots) and reject 4:2:0 for those.
- **Determinism**: For the same input + config, output hash should be stable. Sample periodically: re-run same input, compare content hash; if drift, flag possible race/buffer reuse.

---

## 5) Go Implementation Sketches

### 5.1 Diagnostics Struct
```go
type RawDiag struct {
    LibRawEnabled bool
    Demosaic      string // "AHD","DCB","PPG","unknown"
    OutputBPS     int    // 8|16
    OutputColor   string // "sRGB","AdobeRGB","linear","unknown"
    UseCameraWB   bool
    HalfSize      bool
}

type ResizeDiag struct {
    TargetLongEdge int
    Filter         string
    PreSharpen     struct{Enabled bool; Amount, Radius float64}
    PostSharpen    struct{Enabled bool; Amount, Radius float64}
    Upscale        bool
}

type EncodeDiag struct {
    Format     string // "jpeg","webp","avif"
    Quality    int
    Chroma     string // "420","422","444"
    Progressive bool
    Bytes      int
}

type MetricsDiag struct {
    SSIM        float64
    PSNRdB      float64
    LapVar      float64
    DeltaEMean  float64
    ClippedLow  int
    ClippedHigh int
}

type TimingDiag struct {
    Decode, Orient, Color, Resize, Sharpen, Encode, Store, Total float64
}

type ImageDiag struct {
    ImgID     string
    SourceFmt string
    InputW, InputH int
    HasICC    bool
    ICCDesc   string
    EXIFOrientation int
    Raw       RawDiag
    Pipeline struct{
        OrientationApplied bool
        ColorspaceIn, ColorspaceOut string
        GammaLinearized bool
        Resize          ResizeDiag
        Encode          EncodeDiag
    }
    Metrics   MetricsDiag
    TimingMS  TimingDiag
    FallbackReason string // "none|no_cgo|decode_error|embedded_only"
    Warnings  []string
    Version   struct{ThumbPipeline, LibRaw, Encoder string}
}
```

### 5.2 Instrumentation Hooks
Wrap each stage with timers and diag capture:
```go
func GenerateThumb(ctx context.Context, in image.Image, cfg Config) (out []Thumb, diag ImageDiag, err error) {
    t0 := now()
    defer func() { diag.TimingMS.Total = msSince(t0) }()

    // Decode (or accept decoded input); record libraw params if applicable.
    t := now()
    img, meta, err := decodeWithDiag(...)
    diag.TimingMS.Decode = msSince(t)
    if err != nil { ... }

    // Orientation
    t = now()
    img = applyOrientation(img, meta.Orientation)
    diag.Pipeline.OrientationApplied = true
    diag.TimingMS.Orient = msSince(t)

    // Color / ICC
    t = now()
    img, cdiag := ensureSRGB(img, meta.ICC) // convert if needed
    diag.Pipeline.ColorspaceIn = cdiag.In
    diag.Pipeline.ColorspaceOut = cdiag.Out
    diag.TimingMS.Color = msSince(t)

    // Linear-light resize (optional)
    if cfg.LinearResize {
        diag.Pipeline.GammaLinearized = true
        img = toLinear(img)
    }
    t = now()
    resized := resizeWith(filter, target, img, &diag.Pipeline.Resize)
    diag.TimingMS.Resize = msSince(t)
    if cfg.LinearResize { resized = toSRGB(resized) }

    // Sharpen
    t = now()
    if cfg.PostSharpen.Enabled { resized = unsharp(resized, cfg.PostSharpen) }
    diag.TimingMS.Sharpen = msSince(t)

    // Encode
    t = now()
    buf, enc := encodeWithPolicy(resized, cfg.EncodePolicy)
    diag.Pipeline.Encode = enc
    diag.TimingMS.Encode = msSince(t)

    // Store
    t = now()
    err = storeThumb(buf)
    diag.TimingMS.Store = msSince(t)

    // Metrics vs reference (sampled)
    if sample() {
        diag.Metrics = computeMetrics(resized, referenceOf(resized))
        maybePersistArtifacts(..., diag)
    }

    exportProm(diag) // counters/histograms
    logJSON(diag)    // structured line
    return []Thumb{...}, diag, nil
}
```

### 5.3 Prometheus Wiring (example)
```go
var (
    thumbFallback = promauto.NewCounterVec(prometheus.CounterOpts{Name: "thumb_fallback_total"}, []string{"reason"})
    thumbStageMS  = promauto.NewHistogramVec(prometheus.HistogramOpts{Name: "thumb_stage_ms"}, []string{"stage"})
    thumbBytes    = promauto.NewHistogramVec(prometheus.HistogramOpts{Name: "thumb_bytes"}, []string{"size"})
    thumbSSIM     = promauto.NewSummary(prometheus.SummaryOpts{Name: "thumb_ssim_vs_ref"})
)
```

---

## 6) CI, Tests, and Canaries

- **Golden Orientation Suite**: 16 images covering all EXIF orientations × alpha/no-alpha; tests assert exact pixel match after orientation.
- **ICC/Color Suite**: sRGB, AdobeRGB, Display P3 samples with embedded profiles; assert ΔE mean < threshold after convert→encode→decode roundtrip.
- **RAW Decode Suite**: Selected RAWs where we know expected output properties (bit depth, demosaic choice). Assert `half_size=false`, `output_bps>=16` when configured.
- **Determinism Test**: Repeat full pipeline N times on same input; assert identical content hash and metrics within epsilon.
- **Regression Budget**: In CI, compute SSIM vs reference on a small corpus; enforce lower bound (e.g., P50 SSIM ≥ baseline - 0.005, P95 not worse than -0.01).

---

## 7) Dashboards & Alerts

- **Overview**: Throughput, total time per image, error rate, fallback rate by reason.
- **Quality**: Sampled SSIM/ΔE distributions by size (64/256/512/1024). Trend lines week-over-week.
- **Encoding**: Bytes per thumb size; quality and chroma breakdowns.
- **Raw vs Embedded**: SSIM deltas when fallback_used=true.
- **Hot Spots**: Top offenders (lowest SSIM, highest ΔE) with links to artifacts for repro.

Alert when:
- Fallback rate increases > X% day-over-day.
- Upscale count > 0 (outside explicit mode).
- Orientation double-apply > 0.
- P50 SSIM drops > Y% vs baseline over 24h on canary set.

---

## 8) Operational Controls
- `THUMB_QA_SAMPLE` (e.g., `0.01` = 1%): enable metrics vs reference and artifact capture.
- `THUMB_QA_DIR=/path`: where to store artifacts.
- `THUMB_QA_DISABLE_ARTIFACTS=1`: keep metrics/logs only.
- `THUMB_FORCE_NO_UPSCALE=1`: enforce policy.
- `THUMB_ENCODE_CHROMA=444` for specific MIME categories (screenshots/docs).

---

## 9) Triage Playbook (When Things Look Wrong)
1. **Spike in fallbacks?** Check `reason` labels (no_cgo, decode_error). If `no_cgo`, your binary wasn't built with CGO. If decode_error, inspect artifacts & error logs—bad RAWs? library regression?
2. **SSIM downtrend?** Open top-10 worst artifacts; check for: missing ICC, orientation mistakes, unintended chroma 4:2:0, new sharpening halo, or resize path change.
3. **Upstreams changed?** LibRaw/jpegs updated? Confirm versions in the diag payload.
4. **CPU spiked, quality flat?** A new filter might be active but not helping. Re-run comparison CLI and revert/adjust.
5. **DB bloat?** Bytes histogram shifted; a quality bump (e.g., Q95) slipped in globally. Re-apply quality tiers per size.

---

## 10) Acceptance Criteria (Done = True)
- Per-image **diag JSON** emitted in logs; **Prometheus** metrics exported with key labels.
- **Sampling** produces artifacts and metrics; storage bounded.
- **Guardrails** implemented: no unintended upscales; EXIF applied once; ICC handled; RAW half-size blocked unless opted-in.
- **CI** has golden suites + SSIM floor test; build fails on regression.
- **Dashboards** live with alert rules; on-call playbook documented.

---

## 11) Appendix A — Example Log Line (minified)
```json
{"img_id":"sha256:...","source":{"format":"RAW","input_w":6048,"input_h":4024,"has_icc":true,"exif_orientation":6,"raw":{"libraw_enabled":true,"demosaic":"AHD","output_bps":16,"output_color":"sRGB","use_camera_wb":true,"half_size":false},"fallback_reason":"none"},"pipeline":{"orientation_applied":true,"colorspace_in":"AdobeRGB","colorspace_out":"sRGB","gamma_linearized":true,"resize":{"target_long_edge":1024,"filter":"Lanczos2","pre_sharpen":{"enabled":false},"post_sharpen":{"enabled":true,"amount":0.3,"radius":0.5},"upscale":false},"encode":{"format":"jpeg","quality":90,"chroma":"444","progressive":false,"bytes":201532}},"metrics":{"ssim_vs_ref":0.971,"psnr_vs_ref_db":38.4,"lap_var":232.1,"delta_e_mean":1.2,"clipped_pixels":{"low":41,"high":22}},"timing_ms":{"decode":42.1,"orient":0.3,"color":1.5,"resize":4.1,"sharpen":1.3,"encode":6.8,"store":0.5,"total":58.6},"warnings":[],"version":{"thumb_pipeline":"2.0.0","libraw":"0.21.1","encoder":"libjpeg-turbo"}}
```

---

## 12) Appendix B — Metric Names (Prometheus)
- `thumb_fallback_total{reason}`
- `thumb_stage_ms{stage}`
- `thumb_bytes{size}`
- `thumb_upscale_total`
- `thumb_orientation_double_apply_total`
- `thumb_colorspace_mismatch_total`
- `thumb_quality_bucket_total{q}`
- `thumb_ssim_vs_ref` (sampled)
- `thumb_delta_e_mean` (sampled)

---

## 13) Appendix C — Config Flags (env or CLI)
- `--qa.sample=0.01`
- `--qa.dir=/var/lib/olsen/qa`
- `--qa.disable-artifacts`
- `--thumb.no-upscale`
- `--thumb.linear-resize`
- `--thumb.jpeg.quality-tier=64:80,256:80,512:85,1024:90`
- `--thumb.jpeg.chroma=auto|444|420`
- `--thumb.raw.output-bps=16`
- `--thumb.raw.demosaic=AHD`

---

## 14) Integration Order (Fastest Wins First)
1. Emit **diag JSON + Prom metrics** with stage timings and fallback reasons.
2. Add **no-upscale** guard + orientation & ICC assertions.
3. Turn on **sampling** (1%) with artifacts + SSIM/ΔE vs reference.
4. Add CI **golden tests** and SSIM floor gate.
5. Wire dashboards + alerts.
6. Only then continue with filter/encoder tuning (from v1).

---

## 15) TL;DR for the Engineer
Wire this up now: **metrics, logs, guardrails, sampling**. You’ll immediately know if we’re secretly doing something dumb—then fix that first. Tuning can wait until the pipeline is provably correct.
