type CloudProviderLike = {
  name?: string | null
}

export function orderCloudProviders<T extends CloudProviderLike>(providerList: T[]) {
  const gitProviders: T[] = []
  const remainingProviders: T[] = []

  for (const provider of providerList) {
    if (provider?.name === 'git') {
      gitProviders.push(provider)
      continue
    }
    remainingProviders.push(provider)
  }

  return [...gitProviders, ...remainingProviders]
}
