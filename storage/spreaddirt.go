} else if cAlt > 1000 && cr.chunk.Cells[cLon][cLat][cAlt-1].Material != geom.Air {
	foundInner := false
	innerCheck := []int{
		1, 1,
		1, -1,
		-1, 1,
		-1, -1,
	}
	innerTri := []float32{
		0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,
		-0.5, 0.5, 0.5,

		-0.5, 0.5, 0.5,
		-0.5, -0.5, -0.5,
		0.5, -0.5, 0.5,
	}
	innerTriTcoords := []float32{
		0, 1,
		1, 1,
		1, 0,

		1, 1,
		1, 0,
		0, 0,
	}
	for p := 0; p < len(innerCheck); p += 2 {
		dLon := cLon + innerCheck[p+0]
		dLat := cLat + innerCheck[p+1]
		if dLon < 0 || dLat < 0 || dLon >= cs || dLat >= cs {
			continue
		}
		cell1 := cr.chunk.Cells[dLon][cLat][cAlt]
		cell2 := cr.chunk.Cells[cLon][dLat][cAlt]
		if cell1.Material != geom.Air && cell2.Material != geom.Air {
			pts := make([]float32, len(innerTri))
			for i := 0; i < len(innerTri); i += 3 {
				l := geom.CellLoc{
					Lon: float32(cellIndex.Lon) + float32(innerCheck[p+0])*innerTri[i+0],
					Lat: float32(cellIndex.Lat) + float32(innerCheck[p+1])*innerTri[i+1],
					Alt: float32(cellIndex.Alt) + innerTri[i+2],
				}
				r, theta, phi := planet.CellLocToSpherical(l)
				cart := mgl32.SphericalToCartesian(r, theta, phi)
				pts[i] = cart[0]
				pts[i+1] = cart[1]
				pts[i+2] = cart[2]
			}
			points = append(points, pts...)

			nms := make([]float32, len(pts))
			for i := 0; i < len(pts); i += 9 {
				p1 := mgl32.Vec3{pts[i+0], pts[i+1], pts[i+2]}
				p2 := mgl32.Vec3{pts[i+3], pts[i+4], pts[i+5]}
				p3 := mgl32.Vec3{pts[i+6], pts[i+7], pts[i+8]}
				v1 := p1.Sub(p2)
				v2 := p1.Sub(p3)
				n := v1.Cross(v2).Normalize()
				if n.Dot(p1) < 0 {
					n = n.Mul(-1)
				}
				for j := 0; j < 3; j++ {
					nms[i+3*j+0] = n[0]
					nms[i+3*j+1] = n[1]
					nms[i+3*j+2] = n[2]
				}
			}
			normals = append(normals, nms...)

			tcs := make([]float32, len(innerTriTcoords))
			for i := 0; i < len(innerTriTcoords); i += 2 {
				// 	// tcs[i+0] = float32((i / 2) % 2)
				// 	// tcs[i+1] = float32(1 - ((i / 2) % 2))
				// material := (cLat + cLon) % 7
				material := 1
				tcs[i+0] = (innerTriTcoords[i+0] + float32(material%4)) / 4
				tcs[i+1] = (innerTriTcoords[i+1] + float32(material/4)) / 4
			}
			tcoords = append(tcoords, tcs...)
			foundInner = true
			break
		}
	}

	foundSide := false
	if !foundInner {
		sideCheck := []int{
			1, 0,
			-1, 0,
			0, 1,
			0, -1,
		}
		sideTri := []float32{
			0.5, -0.5, 0.5,
			0.5, 0.5, 0.5,
			-0.5, -0.5, -0.5,

			0.5, 0.5, 0.5,
			-0.5, 0.5, -0.5,
			-0.5, -0.5, -0.5,
		}
		sideTriTcoords := []float32{
			0, 1,
			1, 1,
			0, 0,

			1, 1,
			1, 0,
			0, 0,
		}
		for p := 0; p < len(sideCheck); p += 2 {
			dLon := cLon + sideCheck[p+0]
			dLat := cLat + sideCheck[p+1]
			if dLon < 0 || dLat < 0 || dLon >= cs || dLat >= cs {
				continue
			}
			cell := cr.chunk.Cells[dLon][dLat][cAlt]
			if cell.Material != geom.Air {
				pts := make([]float32, len(sideTri))
				var lonDelta, latDelta int
				lonMul := sideCheck[p+0]
				latMul := sideCheck[p+1]
				if lonMul == 0 {
					lonMul = 1
					lonDelta = 1
					latDelta = 0
				} else {
					latMul = 1
					lonDelta = 0
					latDelta = 1
				}
				for i := 0; i < len(sideTri); i += 3 {
					l := geom.CellLoc{
						Lon: float32(cellIndex.Lon) + float32(lonMul)*sideTri[i+lonDelta],
						Lat: float32(cellIndex.Lat) + float32(latMul)*sideTri[i+latDelta],
						Alt: float32(cellIndex.Alt) + sideTri[i+2],
					}
					r, theta, phi := planet.CellLocToSpherical(l)
					cart := mgl32.SphericalToCartesian(r, theta, phi)
					pts[i] = cart[0]
					pts[i+1] = cart[1]
					pts[i+2] = cart[2]
				}
				points = append(points, pts...)

				nms := make([]float32, len(sideTri))
				for i := 0; i < len(sideTri); i += 9 {
					p1 := mgl32.Vec3{pts[i+0], pts[i+1], pts[i+2]}
					p2 := mgl32.Vec3{pts[i+3], pts[i+4], pts[i+5]}
					p3 := mgl32.Vec3{pts[i+6], pts[i+7], pts[i+8]}
					v1 := p1.Sub(p2)
					v2 := p1.Sub(p3)
					n := v1.Cross(v2).Normalize()
					for j := 0; j < 3; j++ {
						nms[i+3*j+0] = n[0]
						nms[i+3*j+1] = n[1]
						nms[i+3*j+2] = n[2]
					}
				}
				normals = append(normals, nms...)

				tcs := make([]float32, len(sideTriTcoords))
				for i := 0; i < len(sideTriTcoords); i += 2 {
					// 	// tcs[i+0] = float32((i / 2) % 2)
					// 	// tcs[i+1] = float32(1 - ((i / 2) % 2))
					// material := (cLat + cLon) % 7
					material := 1
					tcs[i+0] = (sideTriTcoords[i+0] + float32(material%4)) / 4
					tcs[i+1] = (sideTriTcoords[i+1] + float32(material/4)) / 4
				}
				tcoords = append(tcoords, tcs...)
				foundSide = true
				break
			}
		}
	}

	if !foundInner && !foundSide {
		cornerCheck := []int{
			1, 1,
			-1, 1,
			1, -1,
			-1, -1,
		}
		cornerTri := []float32{
			0.5, 0.5, 0.5,
			0.0, 0.0, -0.5,
			0.5, -0.5, -0.5,

			0.5, 0.5, 0.5,
			-0.5, 0.5, -0.5,
			0.0, 0.0, -0.5,
		}
		cornerTriTcoords := []float32{
			0, 1,
			0.5, 0,
			0, 0,

			1, 1,
			1, 0,
			0.5, 0,
		}
		for p := 0; p < len(cornerCheck); p += 2 {
			dLon := cLon + cornerCheck[p+0]
			dLat := cLat + cornerCheck[p+1]
			if dLon < 0 || dLat < 0 || dLon >= cs || dLat >= cs {
				continue
			}
			cell := cr.chunk.Cells[dLon][dLat][cAlt]
			if cell.Material != geom.Air {
				pts := make([]float32, len(cornerTri))
				for i := 0; i < len(cornerTri); i += 3 {
					l := geom.CellLoc{
						Lon: float32(cellIndex.Lon) + float32(cornerCheck[p+0])*cornerTri[i+0],
						Lat: float32(cellIndex.Lat) + float32(cornerCheck[p+1])*cornerTri[i+1],
						Alt: float32(cellIndex.Alt) + cornerTri[i+2],
					}
					r, theta, phi := planet.CellLocToSpherical(l)
					cart := mgl32.SphericalToCartesian(r, theta, phi)
					pts[i] = cart[0]
					pts[i+1] = cart[1]
					pts[i+2] = cart[2]
				}
				points = append(points, pts...)

				nms := make([]float32, len(cornerTri))
				for i := 0; i < len(cornerTri); i += 9 {
					p1 := mgl32.Vec3{pts[i+0], pts[i+1], pts[i+2]}
					p2 := mgl32.Vec3{pts[i+3], pts[i+4], pts[i+5]}
					p3 := mgl32.Vec3{pts[i+6], pts[i+7], pts[i+8]}
					v1 := p1.Sub(p2)
					v2 := p1.Sub(p3)
					n := v1.Cross(v2).Normalize()
					if n.Dot(p1) < 0 {
						n = n.Mul(-1)
					}
					for j := 0; j < 3; j++ {
						nms[i+3*j+0] = n[0]
						nms[i+3*j+1] = n[1]
						nms[i+3*j+2] = n[2]
					}
				}
				normals = append(normals, nms...)

				tcs := make([]float32, len(cornerTriTcoords))
				for i := 0; i < len(cornerTriTcoords); i += 2 {
					// 	// tcs[i+0] = float32((i / 2) % 2)
					// 	// tcs[i+1] = float32(1 - ((i / 2) % 2))
					// material := (cLat + cLon) % 7
					material := 1
					tcs[i+0] = (cornerTriTcoords[i+0] + float32(material%4)) / 4
					tcs[i+1] = (cornerTriTcoords[i+1] + float32(material/4)) / 4
				}
				tcoords = append(tcoords, tcs...)
				break
			}
		}
	}
