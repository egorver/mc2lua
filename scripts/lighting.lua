local Lighting = game:GetService("Lighting")

-- Убрать глобальные тени
Lighting.GlobalShadows = false

-- Увеличить Ambient освещение
Lighting.Ambient = Color3.fromRGB(200, 200, 200)
Lighting.OutdoorAmbient = Color3.fromRGB(200, 200, 200)

-- Установить время дня (высокое солнце)
Lighting.ClockTime = 14

-- Убрать лучи солнца
local sunRays = Lighting:FindFirstChild("SunRaysEffect")
if sunRays then
	sunRays:Destroy()
end

-- Убрать атмосферу
local atmosphere = Lighting:FindFirstChild("Atmosphere")
if atmosphere then
	atmosphere:Destroy()
end

print("Освещение настроено на равномерное без теней!")
