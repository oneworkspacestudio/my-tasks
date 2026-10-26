import {test,expect} from '@playwright/test';
const url=process.env.APP_E2E_URL,password=process.env.APP_E2E_PASSWORD;
test.setTimeout(45000);
test.use({baseURL:url,viewport:{width:384,height:776}});
test.beforeEach(async({page})=>{
  test.skip(!url||!password,'Requires a disposable native fixture');
  expect(url).toMatch(/^http:\/\/localhost:29\d{3}$/);
  expect((await page.request.post('/api/login',{headers:{Origin:url!},data:{email:'tester@example.test',password}})).status()).toBe(200);
  await page.goto('/');await expect(page.getByRole('heading',{name:'Tasks',exact:true})).toBeVisible();
});
test('reference layout and real task/list interactions',async({page},info)=>{
  const me=await(await page.request.get('/api/me')).json(),headers={Origin:url!,'X-CSRF-Token':me.csrf};
  for(const previous of await(await page.request.get('/api/lists')).json())if(previous.name==='Warsaw')await page.request.delete('/api/lists/'+previous.id,{headers});
  const list=await(await page.request.post('/api/lists',{headers,data:{name:'Warsaw'}})).json();expect(list.id).toBeTruthy();
  let position=0;for(const [title,status] of [['Project 1042 – Review the winter opening hours, confirm availability and agree the delivery date with the team','todo'],['Ear buds','todo'],['Book tickets','done'],['Check travel dates','done'],['Pack charger','done']]){
    expect((await page.request.post('/api/tasks',{headers,data:{title,status,position:position++,listId:list.id,timezone:'Europe/London'}})).status()).toBe(201);
  }
  await page.reload();await page.getByRole('link',{name:'Warsaw',exact:true}).click();
  await expect(page.getByRole('link',{name:'Ear buds',exact:true})).toBeVisible();
  await expect(page.getByRole('button',{name:'Completed (3)',exact:true})).toHaveAttribute('aria-expanded','false');
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  expect(await page.locator('.row--tasks-header').evaluate(e=>getComputedStyle(e).minHeight)).toBe('64px');
  expect(await page.locator('body').evaluate(e=>getComputedStyle(e).backgroundColor)).toBe('rgb(252, 233, 226)');
  await page.screenshot({path:info.outputPath('reference-mobile.png')});
  await page.getByRole('button',{name:'Star Ear buds',exact:true}).click();
  await expect(page.getByRole('button',{name:'Unstar Ear buds',exact:true})).toHaveAttribute('aria-pressed','true');
  await page.getByRole('link',{name:'Starred',exact:true}).click();await expect(page.getByRole('link',{name:'Ear buds',exact:true})).toBeVisible();
  await page.getByRole('link',{name:'Warsaw',exact:true}).click();await page.getByRole('button',{name:'Complete Ear buds',exact:true}).click();
  await expect(page.getByRole('button',{name:'Completed (4)',exact:true})).toBeVisible();await page.getByRole('button',{name:'Completed (4)',exact:true}).click();
  await expect(page.getByRole('link',{name:'Ear buds',exact:true})).toBeVisible();await page.getByRole('button',{name:'Reopen Ear buds',exact:true}).click();await expect(page.getByRole('button',{name:'Complete Ear buds',exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Create a task',exact:true}).click();const editor=page.getByRole('dialog',{name:'Task editor'});await expect(editor).toBeVisible();
  await editor.getByLabel('Title',{exact:true}).fill('A task created in the bottom sheet');await editor.getByLabel('Details',{exact:true}).fill('Synthetic details');await editor.getByRole('button',{name:'Save task',exact:true}).click();
  await expect(page.getByRole('link',{name:'A task created in the bottom sheet',exact:true})).toBeVisible();await page.reload();await expect(page.getByRole('link',{name:'A task created in the bottom sheet',exact:true})).toBeVisible();
  await page.getByRole('link',{name:'Ear buds',exact:true}).click();await expect(editor).toBeVisible();await editor.getByLabel('Title',{exact:true}).fill('Ear buds edited');
  await editor.getByRole('button',{name:'More options',exact:true}).click();await expect(editor.getByLabel('Due date & time',{exact:true})).toBeVisible();await editor.getByRole('button',{name:'Save task',exact:true}).click();await expect(page.getByRole('link',{name:'Ear buds edited',exact:true})).toBeVisible();
  await page.setViewportSize({width:1440,height:1000});await page.screenshot({path:info.outputPath('desktop.png')});
  await page.getByRole('button',{name:'Account menu',exact:true}).click();await page.getByRole('button',{name:'Toggle color theme',exact:true}).click();await expect(page.locator('html')).toHaveAttribute('data-theme','dark');
  await page.setViewportSize({width:320,height:640});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.getByRole('button',{name:'Account menu',exact:true}).click();await page.screenshot({path:info.outputPath('small-dark.png')});
  await page.getByRole('link',{name:'+ New list',exact:true}).click();const manager=page.getByRole('dialog',{name:'Lists & sharing'});await expect(manager).toBeVisible();await manager.getByPlaceholder('New list name').fill('Planning QA');await manager.getByRole('button',{name:'Create list',exact:true}).click();await expect(page.getByRole('heading',{name:'Planning QA',exact:true})).toBeVisible();
  await page.getByRole('button',{name:'List options',exact:true}).click();await expect(page.getByRole('dialog',{name:'List options'})).toBeVisible();await page.getByRole('button',{name:'Close list options',exact:true}).click();
});
test('account binding and private draft clearing',async({page})=>{
  const me=await(await page.request.get('/api/me')).json();expect((await page.request.get('/api/me')).headers()['x-workspace-account']).toBe(me.user.id);
  expect((await page.request.get('/api/me',{headers:{'X-Workspace-Expected-Account':'different-user'}})).status()).toBe(409);
  const email='draft-'+Date.now()+'@example.test';expect((await page.request.post('/api/local-users',{headers:{Origin:url!,'X-CSRF-Token':me.csrf},data:{email,password}})).status()).toBe(201);
  await page.getByRole('button',{name:'Create a task',exact:true}).click();await page.getByLabel('Title',{exact:true}).fill('PRIVATE UNSAVED DRAFT');await page.getByRole('button',{name:'Close',exact:true}).click();
  await page.getByRole('button',{name:'Account menu',exact:true}).click();await page.getByRole('button',{name:'Sign out',exact:true}).click();await page.getByLabel('Email address',{exact:true}).fill(email);await page.getByLabel('Password',{exact:true}).fill(password!);await page.getByRole('button',{name:'Sign in',exact:true}).click();
  await expect(page.getByRole('heading',{name:'Tasks',exact:true})).toBeVisible();await page.getByRole('button',{name:'Create a task',exact:true}).click();await expect(page.getByLabel('Title',{exact:true})).toHaveValue('');await expect(page.getByText('PRIVATE UNSAVED DRAFT',{exact:true})).toHaveCount(0);
});
