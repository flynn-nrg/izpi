# Materials Library for Spectral Rendering

## Overview

The materials library (`internal/materials/`) provides a collection of physically-based materials with precise spectral properties for realistic rendering. Similar to the `lightsources` package, it offers built-in materials that showcase spectral rendering capabilities.

## Architecture

The library provides two main interfaces:

1. **Material Instances**: Direct Material objects that can be used programmatically
2. **Protobuf Definitions**: Material definitions in protobuf format for scene files

## Available Materials

### Porcelain

High-quality ceramic material with realistic spectral reflectance characteristics:

- **Base Color**: White with subtle warm tone
- **Spectral Properties**: 
  - High reflectance (78-93%) across visible spectrum
  - Slightly higher reflectance in red wavelengths (warm appearance)
  - Wavelength-dependent response from 380-750nm in 5nm steps
- **Surface**: Semi-glossy finish (default roughness: 0.15)
- **Subsurface Scattering**: Moderate translucency (strength: 0.05, radius: 0.1)
- **Material Type**: Dielectric (metalness: 0)

#### Variants

- `porcelain`: Default semi-glossy finish
- `porcelain_matte`: Higher roughness (0.4) for matte appearance
- `porcelain_glossy`: Very low roughness (0.05) for glossy appearance

## Usage

### Programmatic Usage

```go
import "github.com/flynn-nrg/izpi/internal/materials"

// Create default porcelain material
porcelain := materials.CreatePorcelain()

// Create custom porcelain with specific parameters
customPorcelain := materials.CreatePorcelainCustom(
    0.2,   // roughness (0.0-1.0)
    0.08,  // SSS strength
    0.15,  // SSS radius
)

// Get material by name from library
mat, ok := materials.GetMaterial("porcelain_glossy")
```

### Scene Integration

For protobuf-based scenes:

```go
import "github.com/flynn-nrg/izpi/internal/materials"

// Create protobuf material definition
porcelainMaterial := materials.CreatePorcelainProtobufMaterial()

// Add to scene materials
protoScene.Materials["Porcelain"] = porcelainMaterial
```

### Custom Parameters

```go
// Create protobuf material with custom settings
customMaterial := materials.CreatePorcelainProtobufMaterialCustom(
    "CustomPorcelain",  // name
    0.25,               // roughness
    0.1,                // SSS strength
    0.2,                // SSS radius
)
```

## Implementation Details

### Spectral Data

The porcelain material uses tabulated spectral reflectance data sampled at 5nm intervals from 380nm to 750nm, matching the CIE standard wavelengths. This ensures accurate color reproduction under different lighting conditions.

### Material Properties

- **Roughness**: Controls surface glossiness
  - 0.0: Mirror-like reflection
  - 0.15: Semi-glossy (default)
  - 1.0: Completely matte
  
- **Subsurface Scattering**: Simulates light penetration
  - Strength: Controls how much light scatters beneath the surface
  - Radius: Controls how far light scatters

### Transport Layer Integration

When used with spectral rendering, the materials are automatically converted to spectral representation by the transport layer. The PBR material's albedo texture is transformed into wavelength-dependent spectral albedo.

## Extending the Library

To add new materials:

1. Define spectral reflectance data (wavelengths and values)
2. Create a constructor function (e.g., `CreateMaterialName()`)
3. Add custom parameter variant (e.g., `CreateMaterialNameCustom()`)
4. Add protobuf generation function if needed
5. Register in `MaterialLibrary` map

Example:

```go
var newMaterialReflectance = []float64{ /* spectral data */ }
var newMaterialWavelengths = []float64{ /* wavelengths */ }

func CreateNewMaterial() material.Material {
    return CreateNewMaterialCustom(defaultRoughness, defaultSSS, defaultRadius)
}

func CreateNewMaterialCustom(roughness, sssStrength, sssRadius float64) material.Material {
    // Create SPD and textures
    // Return PBR material with spectral albedo
}
```

## Example: Stanford Dragon with Porcelain

The Stanford Dragon spectral scene demonstrates the porcelain material:

```go
// In scene definition
dragonMesh.GroupToTransportTrianglesWithMaterial(i, "Porcelain")

// Add to materials
protoScene.Materials["Porcelain"] = materials.CreatePorcelainProtobufMaterial()
```

This produces a realistic white ceramic dragon with subtle warm tones and semi-glossy surface that responds accurately to spectral illumination.

## Future Enhancements

Potential additions to the materials library:

1. **Metal Materials**: Gold, silver, copper with wavelength-dependent reflectance
2. **Stone Materials**: Marble, granite with varied spectral properties
3. **Fabric Materials**: Silk, cotton with translucency
4. **Plastic Materials**: Various polymer types with dispersion
5. **Organic Materials**: Wood, skin with complex scattering

## Technical Notes

- All spectral data uses float64 internally for precision
- Protobuf definitions use float32 for transport efficiency
- Materials implement the full Material interface
- RGB fallback is provided for non-spectral rendering modes
- Materials are thread-safe for parallel rendering

