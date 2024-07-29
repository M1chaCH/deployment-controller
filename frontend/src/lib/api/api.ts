export interface ApiErrorDto {
  message: string;
}

export function isErrorDto(obj: object): obj is ApiErrorDto {
  return (obj as ApiErrorDto).message !== undefined;
}