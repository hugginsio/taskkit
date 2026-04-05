// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package db

import "embed"

//go:embed migration/*.sql
var Migrations embed.FS
