/**
 * Utilities for normalizing protobuf Struct-like data (from Part content.data) into plain objects,
 * and for detecting function-call-shaped payloads.
 * Supports both Buf Value (@bufbuild/protobuf) and legacy Struct-like shapes.
 */

import type { Value } from '@bufbuild/protobuf/wkt'
import { toJson } from '@bufbuild/protobuf'
import { ValueSchema } from '@bufbuild/protobuf/wkt'

/** Value in a Struct-like fieldsMap entry (protobuf JSON representation). */
interface StructLikeValue {
  stringValue?: string
  boolValue?: boolean
  nullValue?: number | null
  numberValue?: number
  structValue?: { fieldsMap?: Record<string, StructLikeValue> | Array<[string, StructLikeValue]> }
  listValue?: { values?: StructLikeValue[] }
}

/** Object that may be a Struct AsObject (fieldsMap as record or array of entries). */
function isStructLike(obj: unknown): obj is Record<string, unknown> & { fieldsMap?: unknown } {
  return (
    typeof obj === 'object' &&
    obj !== null &&
    'fieldsMap' in obj &&
    (Array.isArray((obj as { fieldsMap: unknown }).fieldsMap) ||
      typeof (obj as { fieldsMap: unknown }).fieldsMap === 'object')
  )
}

function normalizeValue(val: StructLikeValue): unknown {
  if (val.stringValue !== undefined) return val.stringValue
  if (val.boolValue !== undefined) return val.boolValue
  if (val.nullValue !== undefined && val.nullValue !== null) return null
  if (val.numberValue !== undefined) return val.numberValue
  if (val.listValue?.values) return val.listValue.values.map(normalizeValue)
  if (val.structValue) return normalizeStructLike(val.structValue)
  return undefined
}

function normalizeStructLike(obj: { fieldsMap?: Record<string, StructLikeValue> | Array<[string, StructLikeValue]> }): Record<string, unknown> {
  const fieldsMap = obj.fieldsMap
  if (!fieldsMap) return {}

  const result: Record<string, unknown> = {}
  const entries: Array<[string, StructLikeValue]> = Array.isArray(fieldsMap)
    ? fieldsMap
    : Object.entries(fieldsMap)

  for (const [key, value] of entries) {
    if (value && typeof value === 'object') {
      result[key] = normalizeValue(value as StructLikeValue)
    }
  }
  return result
}

/**
 * Convert Buf Value to plain object (for @local/a2a-js Part content.data).
 */
export function valueToPlainObject(value: Value | null | undefined): Record<string, unknown> {
  if (!value) return {}
  try {
    const json = toJson(ValueSchema, value)
    if (json && typeof json === 'object' && !Array.isArray(json)) {
      return json as Record<string, unknown>
    }
  } catch {
    // fallback to empty
  }
  return {}
}

/**
 * Normalize Part data into a plain object. Accepts:
 * - Buf Value (from part.content.value when content.case === 'data')
 * - The result of part.getData()?.getData()?.toObject() (Struct AsObject, possibly with fieldsMap)
 * - A raw object wrapped in { data: StructLike }
 * - An object that is already a plain key-value object (returned as-is)
 */
export function normalizeDataPartToObject(raw: unknown): Record<string, unknown> {
  if (raw === null || raw === undefined) return {}

  // Buf Value: use toJson (Part content.data from @local/a2a-js)
  if (typeof raw === 'object' && raw !== null && 'kind' in (raw as object)) {
    return valueToPlainObject(raw as Value)
  }

  // Unwrap { data: ... } if present
  let obj = raw as Record<string, unknown>
  if (typeof obj === 'object' && obj !== null && 'data' in obj && typeof (obj as { data: unknown }).data === 'object') {
    obj = (obj as { data: Record<string, unknown> }).data
  }

  if (!obj || typeof obj !== 'object') return {}

  if (isStructLike(obj)) return normalizeStructLike(obj as { fieldsMap: Record<string, StructLikeValue> | Array<[string, StructLikeValue]> })

  // Already plain object (e.g. from a runtime that normalizes Struct)
  return obj as Record<string, unknown>
}

/**
 * Predicate: true if the object looks like a function call (has string id and name).
 * Use after normalizeDataPartToObject.
 */
export function isFunctionCall(obj: unknown): obj is { id: string; name: string; args?: Record<string, unknown> } {
  if (!obj || typeof obj !== 'object') return false
  const o = obj as Record<string, unknown>
  return typeof o.id === 'string' && o.id.length > 0 && typeof o.name === 'string' && o.name.length > 0
}
