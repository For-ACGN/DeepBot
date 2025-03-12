package deepbot

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/wdvxdr1123/ZeroBot"
)

type txt2Image struct {
	Prompt    string `json:"prompt"`
	NegPrompt string `json:"negative_prompt,omitempty"`

	SamplerName string `json:"sampler_name,omitempty"` // Either SamplerName or SamplerIndex will be used.
	Scheduler   string `json:"scheduler,omitempty"`
	Steps       int    `json:"steps,omitempty"` // How many times to improve the generated image iteratively; higher values take longer; very low values can produce bad results

	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`

	BatchSize  int `json:"batch_size,omitempty"` // How many do you want to simultaneously generate.
	BatchCount int `json:"n_iter,omitempty"`     // How many times do you want to generate.

	DisCFGScale float64 `json:"distilled_cfg_scale,omitempty"`
	CFGScale    float64 `json:"cfg_scale,omitempty"` // Classifier Free Guidance Scale - how strongly the image should conform to prompt - lower values produce more creative results

	EnableHR          bool    `json:"enable_hr,omitempty"`          // Hi-res fix.
	DenoisingStrength float64 `json:"denoising_strength,omitempty"` // Hi-res fix option. Determines how little respect the algorithm should have for image's content. At 0, nothing will change, and at 1 you'll get an unrelated image.
	FirstphaseWidth   int     `json:"firstphase_width,omitempty"`   // Hi-res fix option. Might not work anymore
	FirstphaseHeight  int     `json:"firstphase_height,omitempty"`  // Hi-res fix option. Might not work anymore

	// Hi-res fix option. Multiplier to original width and height.
	//
	// HRScale = 2 will work like this: 384x512 will result in 768x1024
	//
	//  Only HRScale or HRResizeX / HRResizeY will be used
	HRScale float64 `json:"hr_scale,omitempty"`

	// Hi-res fix option. Which Hi-res upscale model will be used.
	//
	//  See: `upscaler` helper package (github.com/Meonako/webui-api/upscaler)
	HRUpscaler string `json:"hr_upscaler,omitempty"`

	// Hi-res fix option. After denoising and upscale, use this amount of steps instead of the amount before denoise and upscale.
	HRSecondPassSteps int `json:"hr_second_pass_steps,omitempty"`

	// Hi-res fix option. The width of the result image
	//
	//  Only HRScale or HRResizeX / HRResizeY will be used
	HRResizeX int `json:"hr_resize_x,omitempty"`

	// Hi-res fix option. The height of the result image
	//
	//  Only HRScale or HRResizeX / HRResizeY will be used
	HRResizeY int `json:"hr_resize_y,omitempty"`

	Seed            int64 `json:"seed,omitempty"` // A value that determines the output of random number generator - if you create an image with same parameters and seed as another image, you'll get the same result
	Subseed         int   `json:"subseed,omitempty"`
	SubseedStrength int   `json:"subseed_strength,omitempty"`
	SeedResizeFromH int   `json:"seed_resize_from_h,omitempty"`
	SeedResizeFromW int   `json:"seed_resize_from_w,omitempty"`

	SamplerIndex string `json:"sampler_index,omitempty"` // Either SamplerName or SamplerIndex will be used.

	RestoreFaces     bool           `json:"restore_faces,omitempty"`
	Tiling           bool           `json:"tiling,omitempty"`
	DoNotSaveSamples bool           `json:"do_not_save_samples,omitempty"`
	DoNotSaveGrid    bool           `json:"do_not_save_grid,omitempty"`
	Eta              float64        `json:"eta,omitempty"`
	SChurn           float64        `json:"s_churn,omitempty"`
	STmax            int            `json:"s_tmax,omitempty"`
	STmin            float64        `json:"s_tmin,omitempty"`
	SNoise           float64        `json:"s_noise,omitempty"`
	OverrideSettings map[string]any `json:"override_settings,omitempty"`

	// Original field was `OverrideSettingsRestoreAfterwards` but since the default value is `true`. it's quite tricky to do this in GO
	//
	//  So I decided to reverse it. This set to true and "override_settings_restore_afterwards": false and vice versa
	DoNotOverrideSettingsRestoreAfterwards bool `json:"override_settings_restore_afterwards"`

	ScriptName string   `json:"script_name,omitempty"`
	ScriptArgs []string `json:"script_args,omitempty"`

	// Original field was `SendImages` but since the default value is `true`. it's quite tricky to do this in GO
	//
	//  So I decided to reverse it. This set to true and "send_images": false and vice versa
	SendImages bool `json:"send_images"`

	// Save image(s) to `outputs` folder where Stable Diffusion Web UI is running
	SaveImages bool `json:"save_images,omitempty"`

	AlwaysOnScripts map[string]any `json:"alwayson_scripts,omitempty"`

	// If true, Will Decode Images after received response from API
	DecodeAfterResult bool `json:"-"`
}

func (bot *DeepBot) onDrawImage(ctx *zero.Ctx) {
	args := textToArgN(ctx.MessageString(), 2)
	if len(args) != 2 {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	prompt := args[1]
	if prompt == " " || prompt == "" {
		bot.sendText(ctx, "非法参数格式")
		return
	}

	sendText(ctx, "正在画图ing", false)

	tr := http.Transport{}
	client := http.Client{
		Transport: &tr,
	}
	defer client.CloseIdleConnections()

	const baseURL = "http://127.0.0.1:7860/sdapi/v1/txt2img"

	arg := txt2Image{
		Prompt:      prompt,
		NegPrompt:   "",
		SamplerName: "Euler a",
		Scheduler:   "Normal",
		Steps:       30,
		Width:       1024,
		Height:      1024,
		BatchSize:   1,
		BatchCount:  1,
		DisCFGScale: 3.5,
		CFGScale:    7,
		SendImages:  true,
		SaveImages:  true,
	}

	data, err := jsonEncode(arg)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL, bytes.NewReader(data))
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	type result struct {
		Images     []string `json:"images"`
		Parameters any      `json:"parameters"`
		Info       string   `json:"info"`
	}
	var results result
	err = jsonDecode(data, &results)
	if err != nil {
		log.Println(err)
		return
	}

	img, err := base64.StdEncoding.DecodeString(results.Images[0])
	if err != nil {
		log.Println(err)
		return
	}

	err = os.WriteFile("F:\\output.png", img, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	sendImage(ctx, img)
}
