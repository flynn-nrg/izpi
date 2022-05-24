# Izpi

![Unit Tests](https://github.com/flynn-nrg/izpi/actions/workflows/test.yml/badge.svg)

A path tracer implemented in Golang built on top of [Peter Shirley's Raytracing books](https://raytracing.github.io).

## Features

* Rendering into a float64 image buffer.
* Direct, indirect and image-based lighting.
* Primitives: Spheres, boxes, rectangles and triangles.
* Wavefront OBJ import.
* Materials: Glass, metal, Lambert, Perlin noise.
* Textures: PNG (LDR) and Radiance (HDR).
* Normal mapping.
* Displacement mapping through sub-texel mesh tessellation.

## Gallery

The [Stanford dragon](https://en.wikipedia.org/wiki/Stanford_dragon)

![The Stanford dragon in a Cornell box](./images/dragon.png "Stanford dragon")

A demonstration of the effect of [displacement mapping](https://en.wikipedia.org/wiki/Displacement_mapping) on a surface using [Bricks078](https://ambientcg.com/view?id=Bricks078) from ambientCG.com, licensed under CC0 1.0 Universal.

![Displacement mapping in a Cornell box](./images/displacement_mapping.png "Displacement mapping")
