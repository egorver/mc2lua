local Lighting = game:GetService("Lighting")

-- Disable global shadows
Lighting.GlobalShadows = false

-- Increase ambient lighting
Lighting.Ambient = Color3.fromRGB(200, 200, 200)
Lighting.OutdoorAmbient = Color3.fromRGB(200, 200, 200)

-- Set time of day (high sun)
Lighting.ClockTime = 14

-- Remove sun rays
local sunRays = Lighting:FindFirstChild("SunRaysEffect")
if sunRays then
	sunRays:Destroy()
end

-- Remove atmosphere
local atmosphere = Lighting:FindFirstChild("Atmosphere")
if atmosphere then
	atmosphere:Destroy()
end

print("Lighting configured: uniform with no shadows!")
